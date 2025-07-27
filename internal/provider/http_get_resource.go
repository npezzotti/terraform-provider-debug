// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"io"
	"net/http"
	"time"

	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &HTTPGetResource{}

func NewHTTPGetResource() resource.Resource {
	return &HTTPGetResource{}
}

type HTTPGetResource struct {
}

type HTTPGetResourceModel struct {
	URL                types.String `tfsdk:"url"`
	Headers            types.Map    `tfsdk:"headers"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	ResponseBody       types.String `tfsdk:"response_body"`
	ResponseStatusCode types.Int64  `tfsdk:"response_status_code"`
	ResponseHeaders    types.Map    `tfsdk:"response_headers"`
}

func (r *HTTPGetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_get"
}

func (r *HTTPGetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "HTTP GET resource",

		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL to perform the HTTP GET request on.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^https?://`), "must start with http:// or https://"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "HTTP headers to include in the request.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						stringvalidator.LengthBetween(1, 256),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must be a valid HTTP header name"),
					),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for the HTTP request in seconds. Defaults to 5 seconds if not set.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(5),
				Validators: []validator.Int64{
					int64validator.Between(1, 60),
				},
			},
			"response_body": schema.StringAttribute{
				MarkdownDescription: "The body of the HTTP response.",
				Computed:            true,
			},
			"response_status_code": schema.Int64Attribute{
				MarkdownDescription: "The HTTP status code of the response.",
				Computed:            true,
			},
			"response_headers": schema.MapAttribute{
				MarkdownDescription: "HTTP headers returned in the response.",
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *HTTPGetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *HTTPGetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HTTPGetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpClient := &http.Client{
		Timeout: time.Duration(data.Timeout.ValueInt64()) * time.Second,
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, data.URL.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Creation Failed",
			"An error occurred while creating the HTTP request: "+err.Error(),
		)
		return
	}

	httpReq.Header = make(http.Header)
	if !data.Headers.IsNull() && !data.Headers.IsUnknown() {
		for k, v := range data.Headers.Elements() {
			if v.IsNull() || v.IsUnknown() {
				continue
			}
			httpReq.Header.Set(k, v.(types.String).ValueString())
		}
	}

	respHTTP, err := httpClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Request Failed",
			"An error occurred while performing the HTTP GET request: "+err.Error(),
		)
		return
	}
	defer respHTTP.Body.Close()

	data.ResponseStatusCode = types.Int64Value(int64(respHTTP.StatusCode))

	headerElements := map[string]string{}
	for k, v := range respHTTP.Header {
		if len(v) > 0 {
			headerElements[k] = v[0] // Use the first value for simplicity
		}
	}

	headers, diags := types.MapValueFrom(ctx, types.StringType, headerElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.ResponseHeaders = headers

	bodyBytes, err := io.ReadAll(respHTTP.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP Response Read Failed",
			"An error occurred while reading the HTTP response body: "+err.Error(),
		)
		return
	}

	if len(bodyBytes) == 0 {
		data.ResponseBody = types.StringNull()
	} else {
		data.ResponseBody = types.StringValue(string(bodyBytes))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HTTPGetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HTTPGetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HTTPGetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HTTPGetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HTTPGetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
