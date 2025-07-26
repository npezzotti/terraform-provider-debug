// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CPUHogDataSource{}

func NewCPUHogDataSource() datasource.DataSource {
	return &CPUHogDataSource{}
}

type CPUHogDataSource struct {
}

type CPUHogDataSourceModel struct {
	NumCores types.Int32  `tfsdk:"num_cores"`
	Duration types.String `tfsdk:"duration"`
}

func (d *CPUHogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cpu_hog"
}

func (d *CPUHogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",

		Attributes: map[string]schema.Attribute{
			"num_cores": schema.Int32Attribute{
				MarkdownDescription: "The number of CPU cores to hog. Leave blank to use all available cores.",
				Optional:            true,
			},
			"duration": schema.StringAttribute{
				MarkdownDescription: "The duration for which to hog the CPU. Defaults to 30 seconds.",
				Optional:            true,
			},
		},
	}
}

func (d *CPUHogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

}

func (d *CPUHogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CPUHogDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	numCores := data.NumCores.ValueInt32()
	if numCores < 0 {
		resp.Diagnostics.AddError(
			"Invalid Number of Cores",
			fmt.Sprintf("The number of cores must be greater than zero, got %d", numCores),
		)
		return
	}

	if numCores == 0 {
		numCores = int32(runtime.NumCPU())
		tflog.Info(ctx, fmt.Sprintf("Using all available CPU cores: %d", numCores))
	} else {
		tflog.Info(ctx, fmt.Sprintf("Using specified number of CPU cores: %d", numCores))
	}

	// Set the number of CPU cores to hog
	if numCores > int32(runtime.NumCPU()) {
		resp.Diagnostics.AddError(
			"Invalid Number of Cores",
			fmt.Sprintf("The number of cores %d exceeds the available CPU cores %d", numCores, runtime.NumCPU()),
		)
		return
	}

	durationStr := data.Duration.ValueString()
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Duration",
			fmt.Sprintf("Failed to parse duration '%s': %v", durationStr, err),
		)
		return
	}

	// Set the maximum number of CPUs to use
	runtime.GOMAXPROCS(int(numCores))

	quit := make(chan struct{})
	var wg sync.WaitGroup
	for i := range numCores {
		wg.Add(1)
		go func(core int32) {
			defer wg.Done()
			tflog.Info(ctx, fmt.Sprintf("Hogging CPU core %d", core))
			hogCPU(quit)
		}(i)
	}

	select {
	case <-time.After(duration):
		tflog.Info(ctx, "CPU hogging duration completed")
		close(quit)
	case <-ctx.Done():
		tflog.Info(ctx, "Context cancelled, stopping CPU hogging")
		close(quit)
	}

	wg.Wait()
}

func hogCPU(stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}
	}
}
