// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &TCPProbeDataSource{}

func NewTCPProbeDataSource() datasource.DataSource {
	return &TCPProbeDataSource{}
}

type TCPProbeDataSource struct {
}

type TCPProbeDataSourceModel struct {
	Host      types.String `tfsdk:"host"`
	Port      types.Int32  `tfsdk:"port"`
	Timeout   types.Int32  `tfsdk:"timeout"`
	UseIPv4   types.Bool   `tfsdk:"use_ipv4"`
	UseIP6    types.Bool   `tfsdk:"use_ipv6"`
	Reachable types.Bool   `tfsdk:"reachable"`
}

func (d *TCPProbeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_probe"
}

func (d *TCPProbeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address to probe.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9.-]+$`), "must be a valid hostname or IP address"),
				},
			},
			"port": schema.Int32Attribute{
				MarkdownDescription: "Port number to probe. Must be between 1 and 65535.",
				Required:            true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			"timeout": schema.Int32Attribute{
				MarkdownDescription: "Timeout for the probe in seconds. Must be between 1 and 60. Defaults to 5 seconds if not set.",
				Optional:            true,
				Validators: []validator.Int32{
					int32validator.Between(1, 60),
				},
			},
			"use_ipv4": schema.BoolAttribute{
				MarkdownDescription: "Use IPv4 for the probe.",
				Optional:            true,
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("use_ipv6")),
				},
			},
			"use_ipv6": schema.BoolAttribute{
				MarkdownDescription: "Use IPv6 for the probe.",
				Optional:            true,
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("use_ipv4")),
				},
			},
			"reachable": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the target is reachable",
				Computed:            true,
			},
		},
	}
}

func (d *TCPProbeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *TCPProbeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TCPProbeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout := data.Timeout.ValueInt32()
	if timeout <= 0 {
		timeout = 5
	}

	dialer := &net.Dialer{
		Timeout: time.Duration(timeout) * time.Second,
	}

	port := strconv.Itoa(int(data.Port.ValueInt32()))
	remoteAddr := net.JoinHostPort(data.Host.ValueString(), port)

	network := "tcp"
	if data.UseIPv4.ValueBool() {
		network += "4"
	}
	if data.UseIP6.ValueBool() {
		network += "6"
	}

	tflog.Info(ctx, "Probing TCP connection", map[string]interface{}{
		"host":     data.Host.ValueString(),
		"port":     port,
		"use_ipv4": data.UseIPv4.ValueBool(),
		"use_ipv6": data.UseIP6.ValueBool(),
		"timeout":  timeout,
	})

	data.Reachable = types.BoolValue(true)
	conn, err := dialer.DialContext(ctx, network, remoteAddr)
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"TCP Probe Failed",
			"Failed to connect to "+data.Host.ValueString()+":"+port+" - "+err.Error(),
		)
		data.Reachable = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
