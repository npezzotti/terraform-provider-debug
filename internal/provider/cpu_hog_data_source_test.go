// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccCPUHogDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:            testAccCPUHogDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{},
			},
		},
	})
}

const testAccCPUHogDataSourceConfig = `
data "debug_cpu_hog" "test" {
  num_cores = 4
	duration = "30s"
}
`
