// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FileContentDataSource{}

func NewFileContentDataSource() datasource.DataSource {
	return &FileContentDataSource{}
}

type FileContentDataSource struct {
}

type FileContentDataSourceModel struct {
	Filename      types.String `tfsdk:"filename"`
	Content       types.String `tfsdk:"content"`
	ContentBase64 types.String `tfsdk:"content_base64"`
	ContentSHA256 types.String `tfsdk:"content_sha256"`
}

func (d *FileContentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_content"
}

func (d *FileContentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{
				MarkdownDescription: "Name of the file to read",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Content of the file",
				Computed:            true,
			},
			"content_base64": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded content of the file",
				Computed:            true,
			},
			"content_sha256": schema.StringAttribute{
				MarkdownDescription: "SHA256 hash of the file content",
				Computed:            true,
			},
		},
	}
}

func (d *FileContentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *FileContentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FileContentDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fileInfo, err := os.Stat(data.Filename.ValueString())
	if err != nil {
		if os.IsNotExist(err) {
			resp.Diagnostics.AddError(
				"File Not Found",
				"The file specified does not exist: "+data.Filename.ValueString(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading File",
				"An error occurred while reading the file: "+err.Error(),
			)
		}
		return
	}

	if fileInfo.IsDir() {
		resp.Diagnostics.AddError(
			"Invalid File Type",
			"The specified path is a directory, not a file: "+data.Filename.ValueString(),
		)
		return
	}

	// Read the file content
	content, err := os.ReadFile(data.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading File",
			"An error occurred while reading the file: "+err.Error(),
		)
		return
	}

	data.Content = types.StringValue(string(content))

	hash := sha256.Sum256(content)
	data.ContentSHA256 = types.StringValue(hex.EncodeToString(hash[:]))

	b64Content := base64.StdEncoding.EncodeToString(content)
	data.ContentBase64 = types.StringValue(b64Content)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
