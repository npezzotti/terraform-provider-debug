// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccOOMKillDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOOMKillDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.debug_oom_kill.test",
						tfjsonpath.New("file_size"),
						knownvalue.Int64Exact(1024),
					),
					statecheck.ExpectKnownValue(
						"data.debug_oom_kill.test",
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(regexp.MustCompile(`^oom-kill-\d+$`)),
					),
				},
			},
		},
	})
}

const testAccOOMKillDataSourceConfig = `
data "debug_oom_kill" "test" {
  memory     = 1024
	block_size = 256
}
`
