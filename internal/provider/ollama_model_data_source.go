// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ollama/ollama/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OllamaModelDataSource{}

func NewOllamaModelDataSource() datasource.DataSource {
	return &OllamaModelDataSource{}
}

// OllamaModelDataSource defines the data source implementation.
type OllamaModelDataSource struct {
	client *api.Client
}

// OllamaModelDataSourceModel describes the data source data model.
type OllamaModelDataSourceModel struct {
	Models []OllamaModel `tfsdk:"models"`
}

func (d *OllamaModelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (d *OllamaModelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of available Ollama models and their details.",

		Attributes: map[string]schema.Attribute{
			"models": schema.ListNestedAttribute{
				Description: "A list of Ollama models, each representing a distinct machine learning model.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the Ollama model.",
							Computed:    true,
						},
						"modified_at": schema.StringAttribute{
							Description: "The date and time when the Ollama model was last modified.",
							Computed:    true,
						},
						"size": schema.Int64Attribute{
							Description: "The size of the Ollama model in bytes.",
							Computed:    true,
						},
						"digest": schema.StringAttribute{
							Description: "A unique identifier (digest) for the version of the Ollama model.",
							Computed:    true,
						},
						"details": schema.SingleNestedAttribute{
							Description: "Detailed attributes of the Ollama model, including format and family.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"format": schema.StringAttribute{
									Description: "The format of the Ollama model (e.g., ONNX, TensorFlow).",
									Optional:    true,
								},
								"family": schema.StringAttribute{
									Description: "The family category to which the Ollama model belongs.",
									Optional:    true,
								},
								"families": schema.ListAttribute{
									Description: "A list of family categories associated with the Ollama model.",
									Optional:    true,
									ElementType: types.StringType,
								},
								"parameter_size": schema.StringAttribute{
									Description: "The size of the parameters within the Ollama model.",
									Optional:    true,
								},
								"quantization_level": schema.StringAttribute{
									Description: "The level of quantization applied to the Ollama model, affecting its precision and size.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *OllamaModelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *OllamaModelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OllamaModelDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rsp, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ollama models, got error: %s", err))
		return
	}

	for _, m := range rsp.Models {
		tflog.Debug(ctx, fmt.Sprintf("m: %#v", m))

		families, familiesDiags := types.ListValueFrom(ctx, types.StringType, m.Details.Families)
		resp.Diagnostics.Append(familiesDiags...)

		data.Models = append(data.Models, OllamaModel{
			Name:       types.StringValue(m.Name),
			ModifiedAt: types.StringValue(m.ModifiedAt.String()),
			Size:       types.Int64Value(m.Size),
			Digest:     types.StringValue(m.Digest),
			Details: OllamaModelDetails{
				Format:            types.StringValue(m.Details.Format),
				Family:            types.StringValue(m.Details.Family),
				Families:          families,
				ParameterSize:     types.StringValue(m.Details.ParameterSize),
				QuantizationLevel: types.StringValue(m.Details.QuantizationLevel),
			},
		})

		tflog.Debug(ctx, fmt.Sprintf("model found: %s", m.Model))
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
