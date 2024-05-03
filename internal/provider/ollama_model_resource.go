package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/ollama/ollama/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &ollamaModelResource{}
	_ resource.ResourceWithConfigure = &ollamaModelResource{}
)

func PullResponseFn(rsp api.ProgressResponse) error {
	tflog.Debug(context.Background(), fmt.Sprintf("ollama Progress response: %#v", rsp))
	return nil
}

// NewOllamaModelResource is a helper function to simplify the provider implementation.
func NewOllamaModelResource() resource.Resource {
	return &ollamaModelResource{}
}

// ollamaModelResource is the resource implementation.
type ollamaModelResource struct {
	client *api.Client
}

func (r *ollamaModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client

}

// Metadata returns the resource type name.
func (r *ollamaModelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (r *ollamaModelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an individual Ollama model, allowing for configuration and tracking of specific models.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique name of the Ollama model. This name is used to identify and manage the model within the system.",
				Required:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "The timestamp when the Ollama model was last modified. This attribute is optional and can be used to track updates.",
				Optional:    true,
			},
			"size": schema.Int64Attribute{
				Description: "The size of the Ollama model in bytes. This attribute is optional and provides information about the model's storage requirements.",
				Optional:    true,
			},
			"digest": schema.StringAttribute{
				Description: "A digest or checksum that uniquely identifies the specific version of the Ollama model. This attribute is optional and helps ensure the integrity of the model.",
				Optional:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ollamaModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OllamaModelResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("model name: %s", plan.Name.String()))

	noStream := false
	ollamaReq := &api.PullRequest{
		Stream: &noStream,
		Name:   plan.Name.ValueString(),
	}
	err := r.client.Pull(ctx, ollamaReq, PullResponseFn)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error pulling model",
			fmt.Sprintf("Could not pull model, unexpected error: %s", err.Error()),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
// Read resource information.
func (r *ollamaModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state OllamaModelResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	ollamaModel, err := r.client.Show(ctx, &api.ShowRequest{Model: state.Name.ValueString()})
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Could not read ollama model %s | %#v", err.Error(), err))

		if apiErr, ok := err.(api.StatusError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Ollama Model",
			"Could not read ollama model "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("ollama show: %#v", ollamaModel))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ollamaModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get current state
	var state OllamaModelResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan OllamaModelResource
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// first delete old model
	tflog.Debug(ctx, fmt.Sprintf("deleting old model: %#v", state.Name.ValueString()))
	err := r.client.Delete(ctx, &api.DeleteRequest{Model: state.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Ollama Model",
			"Could not delete ollama model "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// second pull new model
	noStream := false
	ollamaReq := &api.PullRequest{
		Stream: &noStream,
		Name:   plan.Name.ValueString(),
	}
	if err := r.client.Pull(ctx, ollamaReq, PullResponseFn); err != nil {
		resp.Diagnostics.AddError(
			"Error pulling model",
			fmt.Sprintf("Could not pull model, unexpected error: %s", err.Error()),
		)
		return
	}

	// set new state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ollamaModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state OllamaModelResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, &api.DeleteRequest{Model: state.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Ollama Model",
			"Could not delete ollama model "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}
}
