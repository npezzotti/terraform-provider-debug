// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &OOMKillDataSource{}

func NewOOMKillDataSource() datasource.DataSource {
	return &OOMKillDataSource{}
}

type OOMKillDataSource struct {
}

type OOMKillDataSourceModel struct {
	Memory    types.Int64 `tfsdk:"memory"`
	BlockSize types.Int64 `tfsdk:"block_size"`
}

func (d *OOMKillDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oom_kill"
}

func (d *OOMKillDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Utilize a predetermined amount of memory during the plan phase. " +
			"Useful for testing OOM (Out Of Memory) scenarios.",

		Attributes: map[string]schema.Attribute{
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Amount of memory to allocate.",
				Required:            true,
			},
			"block_size": schema.Int64Attribute{
				MarkdownDescription: "Size of each memory block.",
				Optional:            true,
			},
		},
	}
}

func (d *OOMKillDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

}

func (d *OOMKillDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OOMKillDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	totalBytes := data.Memory.ValueInt64()
	if totalBytes <= 0 && totalBytes != -1 {
		resp.Diagnostics.AddError(
			"Invalid Memory Size",
			"Memory size must be greater than zero or -1 (infinite).",
		)
		return
	}

	blockSize := data.BlockSize.ValueInt64()
	if blockSize < 0 {
		resp.Diagnostics.AddError(
			"Invalid Block Size",
			"Block size must be greater than zero.",
		)
		return
	}

	if blockSize <= 0 {
		blockSize = 100 * 1024 * 1024 // Default to 100MB if not set or invalid
	}

	numBlocksToAllocate := int64(-1)
	var lastBlockSize int64 = blockSize
	if totalBytes != -1 {
		numBlocksToAllocate = totalBytes / blockSize

		// If the last block is not a full block, override the lastBlockSize
		if remainder := totalBytes % blockSize; remainder > 0 {
			numBlocksToAllocate++
			lastBlockSize = remainder
		}
	}

	if totalBytes == -1 {
		tflog.Info(ctx, "Starting infinite memory allocation", map[string]interface{}{
			"block_size": fmt.Sprintf("%d", blockSize),
		})
	} else {
		tflog.Info(ctx, "Starting limited memory allocation", map[string]interface{}{
			"total_bytes": fmt.Sprintf("%d", totalBytes),
			"block_size":  fmt.Sprintf("%d", blockSize),
		})
	}

	totalAllocatedBytes := 0
	memoryHog := make([][]byte, 0)
	for i := 0; ; i++ {
		if numBlocksToAllocate != -1 && int64(i) >= numBlocksToAllocate {
			break
		}

		currBlockSize := blockSize
		if numBlocksToAllocate != -1 && int64(i) == numBlocksToAllocate-1 {
			currBlockSize = lastBlockSize
		}

		// Allocate the memory block.
		block := make([]byte, currBlockSize)
		for j := range block {
			block[j] = byte(j % 256)
		}

		memoryHog = append(memoryHog, block)
		time.Sleep(100 * time.Millisecond) // Simulate some delay in allocation

		totalAllocatedBytes += len(block)

		// Log status every 1000 blocks
		tflog.Debug(ctx, "Allocated memory block", map[string]interface{}{
			"block_index": i,
			"block_size":  fmt.Sprintf("%d", currBlockSize),
			"total_bytes": fmt.Sprintf("%d", totalAllocatedBytes),
		})
	}

	tflog.Info(ctx, "Memory allocation complete", map[string]any{
		"total_bytes":      fmt.Sprintf("%d", totalAllocatedBytes),
		"allocated_blocks": len(memoryHog),
		"block_size":       fmt.Sprintf("%d", blockSize),
	})
}
