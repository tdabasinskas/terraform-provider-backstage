package backstage

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tdabasinskas/go-backstage/backstage"
)

var (
	_ resource.Resource                = &locationResource{}
	_ resource.ResourceWithConfigure   = &locationResource{}
	_ resource.ResourceWithImportState = &locationResource{}
)

// NewLocationResource is a helper function to simplify the provider implementation.
func NewLocationResource() resource.Resource {
	return &locationResource{}
}

// locationResource is the resource implementation.
type locationResource struct {
	client *backstage.Client
}

// locationResourceModel maps the resource schema data.
type locationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Target      types.String `tfsdk:"target"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

const (
	descriptionLocationID          = "Identifier of the location."
	descriptionLocationType        = "Type of the location. Always `url`."
	descriptionLocationTarget      = "Target as a string. Should be a valid URL."
	descriptionLocationLastUpdated = "Timestamp of the last Terraform update of the location."
)

// Metadata returns the data source type name.
func (r *locationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

// Schema defines the schema for the resource.
func (r *locationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this resource to manage Backstage locations. \n\n" +
			"In order for this resource to work, Backstage instance must NOT be running in " +
			"[read-only mode](https://backstage.io/docs/features/software-catalog/configuration#readonly-mode).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, Description: descriptionLocationID, PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			}},
			"type": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: descriptionLocationType, Validators: []validator.String{}, PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			}},
			"target": schema.StringAttribute{Required: true, Description: descriptionLocationTarget,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"last_updated": schema.StringAttribute{Computed: true, Description: descriptionLocationLastUpdated},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *locationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*backstage.Client)
}

// Create registers a new location in Backstage and sets the initial Terraform state.
func (r *locationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan locationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	location, response, err := r.client.Catalog.Locations.Create(ctx, plan.Target.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError("Error creating location",
			fmt.Sprintf("Could not create location, unexpected error: %s", err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Error creating location",
			fmt.Sprintf("Could not create location, unexpected status code: %d", response.StatusCode),
		)
		return
	}

	plan.ID = types.StringValue(location.Location.ID)
	plan.Type = types.StringValue(location.Location.Type)
	plan.Target = types.StringValue(location.Location.Target)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the existing location and refreshes the Terraform state with the latest data.
func (r *locationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state locationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	location, response, err := r.client.Catalog.Locations.GetByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Backstage location",
			fmt.Sprintf("Could not read Backstage location ID %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error reading Backstage location",
			fmt.Sprintf("Could not read Backstage location ID %s, unexpected status code: %d", state.ID.ValueString(), response.StatusCode),
		)
		return
	}

	state.Target = types.StringValue(location.Target)
	state.Type = types.StringValue(location.Type)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *locationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan locationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state locationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the location and removes the Terraform state on success.
func (r *locationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state locationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Catalog.Locations.DeleteByID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Backstage location",
			fmt.Sprintf("Could not delete location, unexpected error: %s", err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("Error deleting Backstage location",
			fmt.Sprintf("Could not delete location, unexpected status code: %d", response.StatusCode),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *locationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
