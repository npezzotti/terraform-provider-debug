// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CommandResource{}

func NewCommandResource() resource.Resource {
	return &CommandResource{}
}

type CommandResource struct {
}

type CommandResourceModel struct {
	Id            types.String `tfsdk:"id"`
	CreateCommand types.List   `tfsdk:"create_command"`
	Stderr        types.String `tfsdk:"stderr"`
	Stdout        types.String `tfsdk:"stdout"`
	ExitCode      types.Int32  `tfsdk:"exit_code"`
}

func (r *CommandResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command"
}

func (r *CommandResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Command resource that executes a command and captures its output.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the resource, used to track state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_command": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Command to be run during the Create operation. Must be a valid command with arguments.",
				Optional:            true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"stderr": schema.StringAttribute{
				MarkdownDescription: "Standard error output from the command.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stdout": schema.StringAttribute{
				MarkdownDescription: "Standard output from the command.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"exit_code": schema.Int32Attribute{
				MarkdownDescription: "Exit code from the command.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CommandResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *CommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CommandResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cmd, diags := data.CreateCommand.ToListValue(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if !cmd.IsNull() && !cmd.IsUnknown() {
		parts := make([]types.String, 0, len(data.CreateCommand.Elements()))
		diags.Append(data.CreateCommand.ElementsAs(ctx, &parts, false)...)

		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		var cmd []string
		for _, part := range parts {
			cmd = append(cmd, part.ValueString())
		}

		idString := strings.Join(cmd, "\x00")
		hash := sha256.Sum256([]byte(idString))
		data.Id = types.StringValue(hex.EncodeToString(hash[:]))

		command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)

		var stderr, stdout strings.Builder
		command.Stderr = &stderr
		command.Stdout = &stdout

		if err := command.Run(); err != nil {
			resp.Diagnostics.AddError(
				"Create Command Failed",
				"An error occurred while executing the create command: "+err.Error(),
			)
			return
		}

		data.Stderr = types.StringValue(stderr.String())
		data.Stdout = types.StringValue(stdout.String())
		data.ExitCode = types.Int32Value(int32(command.ProcessState.ExitCode()))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *CommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *CommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
