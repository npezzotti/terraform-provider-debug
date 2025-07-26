// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FailureResource{}

func NewFailureResource() resource.Resource {
	return &FailureResource{}
}

type FailureResource struct {
}

type FailureResourceModel struct {
	FailOnCreate  types.Bool   `tfsdk:"fail_on_create"`
	FailOnUpdate  types.Bool   `tfsdk:"fail_on_update"`
	FailOnDestroy types.Bool   `tfsdk:"fail_on_destroy"`
	Id            types.String `tfsdk:"id"`
}

func (r *FailureResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failure"
}

func (r *FailureResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Failure resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource. Can be modified to trigger an update. Must be between 10 and 256 characters.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(10, 256),
				},
			},
			"fail_on_create": schema.BoolAttribute{
				MarkdownDescription: "Fail on create",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"fail_on_update": schema.BoolAttribute{
				MarkdownDescription: "Fail on update",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"fail_on_destroy": schema.BoolAttribute{
				MarkdownDescription: "Fail on destroy",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *FailureResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *FailureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FailureResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.FailOnCreate.ValueBool() {
		resp.Diagnostics.AddError(
			"Create Failed",
			"An error occurred while creating the resource.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FailureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FailureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FailureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FailureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.FailOnUpdate.ValueBool() {
		resp.Diagnostics.AddError(
			"Update Failed",
			"An error occurred while updating the resource.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FailureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FailureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.FailOnDestroy.ValueBool() {
		resp.Diagnostics.AddError(
			"Delete Failed",
			"An error occurred while deleting the resource.",
		)
		return
	}
}
