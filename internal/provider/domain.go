package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type OllamaModelResource struct {
	Name       types.String `tfsdk:"name"`
	ModifiedAt types.String `tfsdk:"modified_at"`
	Size       types.Int64  `tfsdk:"size"`
	Digest     types.String `tfsdk:"digest"`
}

type OllamaModel struct {
	Name       types.String       `tfsdk:"name"`
	ModifiedAt types.String       `tfsdk:"modified_at"`
	Size       types.Int64        `tfsdk:"size"`
	Digest     types.String       `tfsdk:"digest"`
	Details    OllamaModelDetails `tfsdk:"details"`
}

type OllamaModelDetails struct {
	Format            types.String `tfsdk:"format" json:"format"`
	Family            types.String `tfsdk:"family" json:"family"`
	Families          types.List   `tfsdk:"families" json:"families"`
	ParameterSize     types.String `tfsdk:"parameter_size" json:"parameter_size"`
	QuantizationLevel types.String `tfsdk:"quantization_level" json:"quantization_level"`
}
