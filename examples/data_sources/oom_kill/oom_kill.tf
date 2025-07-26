terraform {
  cloud {
    hostname     = "dear-flamingo.nathan-pezzotti.sbx.hashidemos.io"
    organization = "nathan-lab"

    workspaces {
      name = "oom-kill"
    }
  }

  required_providers {
    debug = {
      source = "npezzotti/debug"
    }
  }
}

data "debug_oom_kill" "example" {
  memory     = 2 * 1024 * 1024 * 1024 # 2GB
  block_size = 100 * 1024 * 1024      # 100MB
}
