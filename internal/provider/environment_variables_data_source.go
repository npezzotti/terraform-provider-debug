// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EnvDataSource{}

func NewEnvDataSource() datasource.DataSource {
	return &EnvDataSource{}
}

type EnvDataSource struct {
}

type EnvDataSourceModel struct {
	EnvironmentVariables []string          `tfsdk:"environment_variables"`
	Result               map[string]string `tfsdk:"result"`
}

func (d *EnvDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variables"
}

func (d *EnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",

		Attributes: map[string]schema.Attribute{
			"environment_variables": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A list of environment variable names to filter. If empty, all environment variables will be returned.",
				Optional:            true,
			},
			"result": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of environment variables in the run environment.",
				Computed:            true,
			},
		},
	}
}

func (d *EnvDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

}

func (d *EnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	envVars := make(map[string]string)
	// If specific environment variables are provided, only retrieve those.
	if len(data.EnvironmentVariables) > 0 {
		for _, v := range data.EnvironmentVariables {
			val, ok := os.LookupEnv(v)
			if !ok {
				resp.Diagnostics.AddWarning(
					"Environment Variable Not Found",
					fmt.Sprintf("Environment variable '%s' is not set in the current environment.", v),
				)
				continue
			}

			envVars[v] = val
		}
	} else {
		for _, env := range os.Environ() {
			splitVars := strings.SplitN(env, "=", 2)
			if len(splitVars) != 2 {
				resp.Diagnostics.AddWarning(
					"Invalid Environment Variable",
					"Environment variable does not contain a '=' separator: "+env,
				)
				continue
			}

			k, v := splitVars[0], splitVars[1]
			envVars[k] = v
		}
	}

	data.Result = envVars
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
