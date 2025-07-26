terraform {
  # cloud { 
  #   hostname = "dear-flamingo.nathan-pezzotti.sbx.hashidemos.io" 
  #   organization = "nathan-lab" 

  #   workspaces { 
  #     name = "plan-artifact" 
  #   } 
  # } 

  required_providers {
    debug = {
      source = "npezzotti/debug"
    }
  }
}

data "debug_plan_artifact" "example" {
  file_name = "plan-artifact-example"
  file_size = 1024 # 1 MB
}
