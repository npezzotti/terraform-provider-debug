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

func TestAccPlanArtifactDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPlanArtifactDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.debug_plan_artifact.test",
						tfjsonpath.New("file_size"),
						knownvalue.Int64Exact(1024),
					),
					statecheck.ExpectKnownValue(
						"data.debug_plan_artifact.test",
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(regexp.MustCompile(`^plan-artifact-\d+$`)),
					),
				},
			},
		},
	})
}

const testAccPlanArtifactDataSourceConfig = `
data "debug_plan_artifact" "test" {
  file_size = 1024
}
`
