// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SystemInfoDataSource{}

func NewSystemInfoDataSource() datasource.DataSource {
	return &SystemInfoDataSource{}
}

type SystemInfoDataSource struct {
}

type SystemInfoDataSourceModel struct {
	Hostname     types.String `tfsdk:"hostname"`
	OS           types.String `tfsdk:"os"`
	PlatformInfo types.Object `tfsdk:"platform_info"`
	NumCPUs      types.Int64  `tfsdk:"num_cpus"`
	MemoryTotal  types.Int64  `tfsdk:"memory_total"`
	DiskInfo     types.Object `tfsdk:"disk_info"`
	ProcInfo     types.Object `tfsdk:"proc_info"`
	Home         types.String `tfsdk:"home"`
	Path         types.String `tfsdk:"path"`
	WorkingDir   types.String `tfsdk:"working_dir"`
}

type ProcInfo struct {
	Uid  types.Int64 `tfsdk:"uid"`
	Gid  types.Int64 `tfsdk:"gid"`
	PID  types.Int64 `tfsdk:"pid"`
	PPID types.Int64 `tfsdk:"ppid"`
}

type DiskInfo struct {
	Total types.Int64 `tfsdk:"total"` // Total disk space in bytes
	Used  types.Int64 `tfsdk:"used"`  // Used disk space in bytes
	Free  types.Int64 `tfsdk:"free"`  // Free disk space in bytes
}

type PlatformInfo struct {
	Platform        string `tfsdk:"platform"`
	PlatformFamily  string `tfsdk:"platform_family"`
	PlatformVersion string `tfsdk:"platform_version"`
	KernelVersion   string `tfsdk:"kernel_version"`
	KernelArch      string `tfsdk:"kernel_arch"`
}

func (d *SystemInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_info"
}

func (d *SystemInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "System information data source",

		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname of the system",
				Computed:            true,
			},
			"os": schema.StringAttribute{
				MarkdownDescription: "Operating system name",
				Computed:            true,
			},
			"platform_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Platform information including platform, family, version, kernel version, and architecture",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"platform":         schema.StringAttribute{Computed: true},
					"platform_family":  schema.StringAttribute{Computed: true},
					"platform_version": schema.StringAttribute{Computed: true},
					"kernel_version":   schema.StringAttribute{Computed: true},
					"kernel_arch":      schema.StringAttribute{Computed: true},
				},
			},
			"proc_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Process information including UID, GID, PID, and PPID",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"uid":  schema.Int64Attribute{Computed: true},
					"gid":  schema.Int64Attribute{Computed: true},
					"pid":  schema.Int64Attribute{Computed: true},
					"ppid": schema.Int64Attribute{Computed: true},
				},
			},
			"num_cpus": schema.Int64Attribute{
				MarkdownDescription: "Number of CPU cores available on the system",
				Computed:            true,
			},
			"memory_total": schema.Int64Attribute{
				MarkdownDescription: "Total memory available on the system in bytes",
				Computed:            true,
			},
			"disk_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Disk information including total, used, and free space",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"total": schema.Int64Attribute{
						MarkdownDescription: "Total disk space in bytes",
						Computed:            true,
					},
					"used": schema.Int64Attribute{
						MarkdownDescription: "Used disk space in bytes",
						Computed:            true,
					},
					"free": schema.Int64Attribute{
						MarkdownDescription: "Free disk space in bytes",
						Computed:            true,
					},
				},
			},
			"home": schema.StringAttribute{
				MarkdownDescription: "Home directory of the current user",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Path to the provider's root directory",
				Computed:            true,
			},
			"working_dir": schema.StringAttribute{
				MarkdownDescription: "Current working directory of the process",
				Computed:            true,
			},
		},
	}
}

func (d *SystemInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
}

func (d *SystemInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemInfoDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	info, err := host.InfoWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get system info",
			"An unexpected error occurred while getting system information: "+err.Error())
		return
	}

	data.Hostname = types.StringValue(info.Hostname)
	data.OS = types.StringValue(info.OS)

	var diags diag.Diagnostics
	data.PlatformInfo, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"platform":         types.StringType,
		"platform_family":  types.StringType,
		"platform_version": types.StringType,
		"kernel_version":   types.StringType,
		"kernel_arch":      types.StringType,
	}, PlatformInfo{
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
	})

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	counts, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get CPU counts",
			"An unexpected error occurred while getting CPU counts: "+err.Error())
		return
	}

	data.NumCPUs = types.Int64Value(int64(counts))

	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get memory info",
			"An unexpected error occurred while getting memory information: "+err.Error())
		return
	}

	data.MemoryTotal = types.Int64Value(int64(memInfo.Total))

	workingDir, err := os.Getwd()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get working directory",
			"An unexpected error occurred while getting the working directory: "+err.Error())
		return
	}

	data.WorkingDir = types.StringValue(workingDir)

	diskInfo, err := disk.UsageWithContext(ctx, workingDir)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get disk info",
			"An unexpected error occurred while getting disk information: "+err.Error())
		return
	}

	diskData := DiskInfo{
		Total: types.Int64Value(int64(diskInfo.Total)),
		Used:  types.Int64Value(int64(diskInfo.Used)),
		Free:  types.Int64Value(int64(diskInfo.Free)),
	}

	data.DiskInfo, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"total": types.Int64Type,
		"used":  types.Int64Type,
		"free":  types.Int64Type,
	}, diskData)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	path, ok := os.LookupEnv("PATH")
	if ok {
		data.Path = types.StringValue(path)
	} else {
		resp.Diagnostics.AddError("Unable to get PATH",
			"Failed to retrieve the PATH environment variable.")
		return
	}

	data.ProcInfo, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"uid":  types.Int64Type,
		"gid":  types.Int64Type,
		"pid":  types.Int64Type,
		"ppid": types.Int64Type,
	}, ProcInfo{
		Uid:  types.Int64Value(int64(os.Getuid())),
		Gid:  types.Int64Value(int64(os.Getgid())),
		PID:  types.Int64Value(int64(os.Getpid())),
		PPID: types.Int64Value(int64(os.Getppid())),
	})

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get home directory",
			"An unexpected error occurred while getting the home directory: "+err.Error())
		return
	}
	data.Home = types.StringValue(homeDir)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
