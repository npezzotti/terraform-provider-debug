terraform {
  cloud {
    hostname     = "dear-flamingo.nathan-pezzotti.sbx.hashidemos.io"
    organization = "nathan-lab"

    workspaces {
      name = "cpu-hog"
    }
  }

  required_providers {
    debug = {
      source = "npezzotti/debug"
    }
  }
}

data "debug_cpu_hog" "example" {
  num_cores = 4
  duration  = "30s"
}
