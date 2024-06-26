package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &resourceResource{}
	_ resource.ResourceWithConfigure = &resourceResource{}
)

func NewResource() resource.Resource {
	return &resourceResource{}
}

type resourceResource struct {
	providerConfig *httpclientProviderConfig
}

type resourceStateModel struct {
	Variables   types.Map    `tfsdk:"variables"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (resource *resourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (resource *resourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Map of variables to replace in the URL",
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time when this resource was last updated",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.providerConfig = req.ProviderData.(*httpclientProviderConfig)
}

func (resource *resourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Load resource config from plan
	var resourceData resourceStateModel
	resp.Diagnostics.Append(
		req.Plan.Get(ctx, &resourceData)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make the API call
	err := makeApiRequest(
		ctx,
		*resource,
		resourceData,
		resource.providerConfig.CreateUrl.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed making request", err.Error())
	}

	// Add data to the resource
	resourceData.LastUpdated = types.StringValue(
		time.Now().Format(time.RFC3339),
	)

	// Save resource state
	resp.Diagnostics.Append(
		resp.State.Set(ctx, resourceData)...,
	)
}

func (resource *resourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load resource config from state
	var resourceData resourceStateModel
	resp.Diagnostics.Append(
		req.State.Get(ctx, &resourceData)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make the API call
	err := makeApiRequest(
		ctx,
		*resource,
		resourceData,
		resource.providerConfig.ReadUrl.ValueString(),
	)
	if err != nil {
		// resp.Diagnostics.AddError("Failed making request", err.Error())
		resp.State.RemoveResource(ctx)
		return
	}

	// Update resource data
	resourceData.LastUpdated = types.StringValue(
		time.Now().Format(time.RFC3339),
	)

	// Save resource state
	resp.Diagnostics.Append(
		resp.State.Set(ctx, resourceData)...,
	)
}

func (resource *resourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Load resource data from state
	var resourceData resourceStateModel
	resp.Diagnostics.Append(
		req.Plan.Get(ctx, &resourceData)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}
	resourceData.LastUpdated = types.StringValue(
		time.Now().Format(time.RFC3339),
	)
}

func (resource *resourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load resource config from state
	var resourceData resourceStateModel
	resp.Diagnostics.Append(
		req.State.Get(ctx, &resourceData)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make the API call
	err := makeApiRequest(
		ctx,
		*resource,
		resourceData,
		resource.providerConfig.DeleteUrl.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed making request", err.Error())
	}
}

func makeApiRequest(
	ctx context.Context,
	resource resourceResource,
	resourceData resourceStateModel,
	url string,
) error {
	// Replace placeholders in url with variable values
	variables := map[string]string{}
	resourceData.Variables.ElementsAs(ctx, &variables, false)
	for variable, value := range variables {
		url = strings.ReplaceAll(url, "{"+variable+"}", value)
	}

	// Send the request to the API
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	apiResponse, err := resource.providerConfig.client.Do(request)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(apiResponse.Body)
	if err != nil {
		return err
	}
	if apiResponse.StatusCode != 200 {
		return fmt.Errorf(
			"request unsuccessful\nstatus code: %d\nbody: %s",
			apiResponse.StatusCode,
			body,
		)
	}

	// Everything went ok
	return nil
}
