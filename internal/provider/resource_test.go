package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUrlBased(t *testing.T) {
	mockApiObjects := map[string]string{}
	apiServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				t.Logf("API Received \"%s\" request to URI \"%s\"", r.Method, r.RequestURI)

				if r.Method != "GET" {
					t.Fatal("Bad request method", r.Method)
				}

				r.ParseForm()
				uriParts := strings.Split(r.RequestURI, "/")
				t.Log("Parsed action:", uriParts[1])

				switch uriParts[1] {
				case "create":
					mockApiObjects[uriParts[2]] = uriParts[3]
				case "read":
					value, success := mockApiObjects[uriParts[2]]
					if success {
						w.Write([]byte(value))
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				case "delete":
					delete(mockApiObjects, uriParts[2])
				default:
					t.Fatal("Bad request URI (action):", uriParts[1])
				}
			},
		),
	)
	defer apiServer.Close()

	terraformCode := fmt.Sprintf(
		`
			provider "httpclient" {
				create_url = "%[1]s/create/{name}/{value}"
				read_url = "%[1]s/read/{name}"
				delete_url = "%[1]s/delete/{name}"
			}

			resource "httpclient_resource" "test-resource" {
				variables = {
					name = "name-1"
					value = "value-1"
				}
			}
		`,
		apiServer.URL,
	)

	testCase := resource.TestCase{
		ProtoV6ProviderFactories: providerFactory,
		PreCheck: func() {
			response, _ := http.Get(apiServer.URL + "/read/name-1")
			if response.StatusCode == 200 {
				t.Fatal("Object already exists in the API")
			}
		},
		Steps: []resource.TestStep{
			{
				Config: terraformCode,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if resource was created in the API
					func(s *terraform.State) error {
						response, _ := http.Get(apiServer.URL + "/read/name-1")
						// Check if object exists
						if response.StatusCode != 200 {
							t.Fatal("Object not found in the API")
						}
						// Check if it has the correct content
						responseString, err := io.ReadAll(response.Body)
						if err != nil {
							t.Fatal("Error reading response: ", err)
						}
						if string(responseString) != "value-1" {
							t.Fatal("Object content not found in the API response")
						}
						// Checks passed
						return nil
					},
					// Check resource is present in the state
					resource.TestCheckResourceAttr("httpclient_resource.test-resource", "variables.name", "name-1"),
					resource.TestCheckResourceAttr("httpclient_resource.test-resource", "variables.value", "value-1"),
				),
			},
		},
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckNoResourceAttr("httpclient_resource.test-resource", "variables.name"),
		),
	}

	resource.Test(t, testCase)
}
