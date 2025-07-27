// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &DebugProvider{}
var _ provider.ProviderWithFunctions = &DebugProvider{}
var _ provider.ProviderWithEphemeralResources = &DebugProvider{}

type DebugProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type DebugProviderModel struct {
}

func (p *DebugProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "debug"
	resp.Version = p.version
}

func (p *DebugProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{}}
}

func (p *DebugProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DebugProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *DebugProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFailureResource,
		NewSleepResource,
		NewCommandResource,
		NewHTTPGetResource,
	}
}

func (p *DebugProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *DebugProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPlanArtifactDataSource,
		NewOOMKillDataSource,
		NewCPUHogDataSource,
		NewEnvDataSource,
		NewDNSLookupDataSource,
		NewTCPProbeDataSource,
		NewFileContentDataSource,
		NewFailureDataSource,
		NewSystemInfoDataSource,
	}
}

func (p *DebugProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DebugProvider{
			version: version,
		}
	}
}
