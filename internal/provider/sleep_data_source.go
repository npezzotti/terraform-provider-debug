// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &SleepDataSource{}

func NewSleepDataSource() datasource.DataSource {
	return &SleepDataSource{}
}

type SleepDataSource struct {
}

type SleepDataSourceModel struct {
	Duration types.String `tfsdk:"duration"`
}

func (d *SleepDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sleep"
}

func (d *SleepDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The sleep data source allows you to pause execution for a specified duration. " +
			"This can be useful for simulating long-running operations or for inspecting the run environment during a Terraform plan.",

		Attributes: map[string]schema.Attribute{
			"duration": schema.StringAttribute{
				MarkdownDescription: "Duration to sleep, e.g. '30s' for 30 seconds.",
				Required:            true,
			},
		},
	}
}

func (d *SleepDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *SleepDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SleepDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	duration, err := time.ParseDuration(data.Duration.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Duration",
			"Could not parse duration: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Sleeping for duration", map[string]interface{}{
		"duration": duration.String(),
	})

	if err := sleep(ctx, duration); err != nil {
		resp.Diagnostics.AddError(
			"Sleep Error",
			"An error occurred while sleeping: "+err.Error(),
		)
		return
	}
}
