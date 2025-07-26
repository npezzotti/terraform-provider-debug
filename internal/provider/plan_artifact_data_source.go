// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &PlanArtifactDataSource{}

func NewPlanArtifactDataSource() datasource.DataSource {
	return &PlanArtifactDataSource{}
}

type PlanArtifactDataSource struct {
}

type PlanArtifactDataSourceModel struct {
	FileSize types.Int64  `tfsdk:"file_size"`
	FileName types.String `tfsdk:"file_name"`
	Id       types.String `tfsdk:"id"`
}

func (d *PlanArtifactDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plan_artifact"
}

func (d *PlanArtifactDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artifact with configurable size generated during the plan phase. " +
			"Useful for testing issues during the upload plan filesystem phase of a run.",

		Attributes: map[string]schema.Attribute{
			"file_size": schema.Int64Attribute{
				MarkdownDescription: "Size of the artifact file in bytes.",
				Required:            true,
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "Name of the generated file.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Sha256 hash of the generated file.",
				Computed:            true,
			},
		},
	}
}

func (d *PlanArtifactDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

}

func (d *PlanArtifactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PlanArtifactDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if existingFile, err := os.Stat(data.FileName.ValueString()); err != nil {
		if !os.IsNotExist(err) {
			resp.Diagnostics.AddError(
				"Unable to read existing file",
				fmt.Sprintf("An error occurred while checking the existing file: %s", err),
			)
			return
		}
		tflog.Info(ctx, fmt.Sprintf("File %s does not exist, creating new one.", data.FileName.ValueString()))
	} else {
		tflog.Info(ctx, fmt.Sprintf("File %s already exists, using it.", existingFile.Name()))
		data.FileSize = types.Int64Value(existingFile.Size())
		data.FileName = types.StringValue(filepath.Base(existingFile.Name()))

		fh, err := os.Open(existingFile.Name())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to open existing file",
				fmt.Sprintf("An error occurred while opening the existing file: %s", err),
			)
			return
		}
		hash, err := hashFile(fh)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to hash existing file",
				fmt.Sprintf("An error occurred while hashing the existing file: %s", err),
			)
			return
		}

		data.Id = types.StringValue(hash)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	var (
		chunkSize int64 = 64 * 1024 // 64 KiB
		fileSize  int64 = data.FileSize.ValueInt64()
	)

	if fileSize <= 0 {
		resp.Diagnostics.AddError(
			"Invalid File Size",
			"File size must be greater than zero.",
		)
		return
	}

	fh, err := os.Create(data.FileName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create file",
			fmt.Sprintf("An error occurred while creating the file: %s", err),
		)
		return
	}
	defer fh.Close()

	chunk := make([]byte, chunkSize)
	var written int64
	for written < fileSize {
		bytes := chunkSize
		if (fileSize - written) < chunkSize {
			bytes = fileSize - written
		}

		_, err := rand.Read(chunk[:bytes])
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to generate random data",
				fmt.Sprintf("An error occurred while generating random data: %s", err),
			)
			return
		}

		n, err := fh.Write(chunk[:bytes])
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to write to file",
				fmt.Sprintf("An error occurred while writing to the file: %s", err),
			)
			return
		}

		written += int64(n)
	}

	tflog.Info(ctx, fmt.Sprintf("Wrote %d bytes to file %s", written, fh.Name()))

	data.FileSize = types.Int64Value(written)
	data.FileName = types.StringValue(filepath.Base(fh.Name()))

	// Reset file pointer to the beginning before hashing
	if _, err := fh.Seek(0, io.SeekStart); err != nil {
		resp.Diagnostics.AddError(
			"Unable to seek file",
			fmt.Sprintf("An error occurred while seeking the file before hashing: %s", err),
		)
		return
	}

	hash, err := hashFile(fh)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to hash file",
			fmt.Sprintf("An error occurred while hashing the file: %s", err),
		)
		return
	}

	data.Id = types.StringValue(hash)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func hashFile(fh *os.File) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fh); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
