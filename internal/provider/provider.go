package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &httpclientProvider{}
)

type httpclientProvider struct {
	version string
}

type httpclientProviderConfig struct {
	client    http.Client
	CreateUrl types.String `tfsdk:"create_url"`
	ReadUrl   types.String `tfsdk:"read_url"`
	DeleteUrl types.String `tfsdk:"delete_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &httpclientProvider{
			version: version,
		}
	}
}

func (p *httpclientProvider) Metadata(ctx context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "httpclient"
	resp.Version = p.version
}

func (p *httpclientProvider) Schema(ctx context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage resources through any API with Terraform",
		Attributes: map[string]schema.Attribute{
			"create_url": schema.StringAttribute{
				Description: "The URL to call to create the resource",
				Required:    true,
			},
			"read_url": schema.StringAttribute{
				Description: "The URL to call to check if resource exists",
				Required:    true,
			},
			"delete_url": schema.StringAttribute{
				Description: "The URL to call to delete the resource",
				Required:    true,
			},
		},
	}
}

func (p *httpclientProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var providerConfig httpclientProviderConfig
	req.Config.Get(ctx, &providerConfig)

	providerConfig.client = http.Client{}
	resp.ResourceData = &providerConfig
	resp.DataSourceData = &providerConfig
}

func (p *httpclientProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *httpclientProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResource,
	}
}
