terraform {
  required_providers {
    httpclient = {
      source = "registry.terraform.io/kvaidas/httpclient"
    }
  }
}

provider "httpclient" {
  create_url = "https://httpbin.org/get"
  read_url = "https://httpbin.org/get"
  delete_url = "https://httpbin.org/get"
}

resource "httpclient_resource" "test-resource" {
  variables = {
    some_variable = "some_value"
  }
}
