// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SleepResource{}

func NewSleepResource() resource.Resource {
	return &SleepResource{}
}

type SleepResource struct {
}

type SleepResourceModel struct {
	Id              timetypes.RFC3339 `tfsdk:"id"`
	Duration        types.String      `tfsdk:"duration"`
	UpdateDuration  types.String      `tfsdk:"update_duration"`
	DestroyDuration types.String      `tfsdk:"destroy_duration"`
}

func (r *SleepResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sleep"
}

func (r *SleepResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The sleep resource allows you to pause execution for a specified duration during create, update, or destroy operations. " +
			"This can be useful for simulating long-running operations or for inspecting the run environment during a Terraform run.",

		Attributes: map[string]schema.Attribute{
			"duration": schema.StringAttribute{
				MarkdownDescription: "Duration to sleep before completing the create operation. Must be a valid duration string (e.g., '5s', '1m').",
				Required:            true,
			},
			"update_duration": schema.StringAttribute{
				MarkdownDescription: "Duration to sleep before completing the update operation. Must be a valid duration string (e.g., '5s', '1m').",
				Optional:            true,
			},
			"destroy_duration": schema.StringAttribute{
				MarkdownDescription: "Duration to sleep before completing the destroy operation. Must be a valid duration string (e.g., '5s', '1m').",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The time when the resource was created or last updated, in RFC3339 format.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SleepResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *SleepResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SleepResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	durString := data.Duration.ValueString()
	duration, err := time.ParseDuration(durString)
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

	err = sleep(ctx, duration)
	if err != nil {
		resp.Diagnostics.AddError(
			"Sleep Failed",
			"An error occurred while sleeping: "+err.Error(),
		)
		return
	}

	data.Id = timetypes.NewRFC3339TimeValue(time.Now().UTC())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SleepResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *SleepResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SleepResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	durString := state.UpdateDuration.ValueString()
	if durString != "" {
		duration, err := time.ParseDuration(durString)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Update Duration",
				"Could not parse update_duration: "+err.Error(),
			)
			return
		}

		tflog.Info(ctx, "Sleeping for duration", map[string]interface{}{
			"duration": duration.String(),
		})

		err = sleep(ctx, duration)
		if err != nil {
			resp.Diagnostics.AddError(
				"Sleep Failed",
				"An error occurred while sleeping: "+err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SleepResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SleepResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	durString := data.DestroyDuration.ValueString()
	if durString == "" {
		tflog.Info(ctx, "No destroy_duration specified, skipping sleep")
		return
	}

	duration, err := time.ParseDuration(durString)
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

	err = sleep(ctx, duration)
	if err != nil {
		resp.Diagnostics.AddError(
			"Sleep Failed",
			"An error occurred while sleeping: "+err.Error(),
		)
		return
	}
}

func sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
