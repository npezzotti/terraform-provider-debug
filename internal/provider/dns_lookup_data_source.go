// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DNSLookupDataSource{}

func NewDNSLookupDataSource() datasource.DataSource {
	return &DNSLookupDataSource{}
}

// DNSLookupDataSource defines the data source implementation.
type DNSLookupDataSource struct {
}

// DNSLookupDataSourceModel describes the data source data model.
type DNSLookupDataSourceModel struct {
	Hostname types.String `tfsdk:"hostname"`
	Result   types.List   `tfsdk:"result"`
}

func (d *DNSLookupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_lookup"
}

func (d *DNSLookupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname to look up.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"result": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Result of the DNS lookup.",
				Computed:            true,
			},
		},
	}
}

func (d *DNSLookupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *DNSLookupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSLookupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	ips, err := net.LookupIP(data.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to look up IP addresses",
			fmt.Sprintf("Could not look up IP addresses for hostname %s: %s", data.Hostname.String(), err),
		)
		return
	}

	var res []string
	for _, ip := range ips {
		res = append(res, ip.String())
	}

	elements := make([]types.String, 0, len(data.Result.Elements()))
	for _, ip := range res {
		elements = append(elements, types.StringValue(ip))
	}

	listValue, diags := types.ListValueFrom(ctx, types.StringType, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Result = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
