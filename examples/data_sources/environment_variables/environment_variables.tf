terraform {
  cloud {
    hostname     = "dear-flamingo.nathan-pezzotti.sbx.hashidemos.io"
    organization = "nathan-lab"

    workspaces {
      name = "env"
    }
  }

  required_providers {
    debug = {
      source = "npezzotti/debug"
    }
  }
}

data "debug_environment_variables" "example" {
  environment_variables = ["HOME"]
}

output "environment_variables" {
  value = data.debug_environment_variables.example.result
}
