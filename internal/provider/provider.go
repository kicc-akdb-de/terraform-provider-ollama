// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/ollama/ollama/api"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	_ "github.com/ollama/ollama/api"
)

// Ensure OllamaProvider satisfies various provider interfaces.
var _ provider.Provider = &OllamaProvider{}
var _ provider.ProviderWithFunctions = &OllamaProvider{}

// OllamaProvider defines the provider implementation.
type OllamaProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OllamaProviderModel describes the provider data model.
type OllamaProviderModel struct {
	Host types.String `tfsdk:"host"`
}

func (p *OllamaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ollama"
	resp.Version = p.version
}

func (p *OllamaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Ollama host",
				Required:    true,
			},
		},
	}
}

func (p *OllamaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config OllamaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("OLLAMA_HOST")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing HashiCups API Host",
			"The provider cannot create the HashiCups API client as there is a missing or empty value for the HashiCups API host. "+
				"Set the host value in the configuration or use the HASHICUPS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	os.Setenv("OLLAMA_HOST", host) // TODO change this when ollama sdk changes to not just use 'from env'

	// Example client configuration for config sources and resources
	client, err := api.ClientFromEnvironment()

	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Error creating ollama client",
			"The provider cannot create the ollama API client as there is a missing or empty value for the OLLAMA_HOST or ollama host. "+
				"Set the host value in the configuration or use the OLLAMA_HOST environment variable. "+
				"If either is already set, ensure the value is not empty or broken.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OllamaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOllamaModelResource,
	}
}

func (p *OllamaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOllamaModelDataSource,
	}
}

func (p *OllamaProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OllamaProvider{
			version: version,
		}
	}
}
