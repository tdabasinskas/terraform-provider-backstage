package backstage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tdabasinskas/go-backstage/backstage"
)

var (
	_ datasource.DataSource              = &apiDataSource{}
	_ datasource.DataSourceWithConfigure = &apiDataSource{}
)

// NewApiDataSource is a helper function to simplify the provider implementation.
func NewApiDataSource() datasource.DataSource {
	return &apiDataSource{}
}

// apiDataSource is the data source implementation.
type apiDataSource struct {
	client *backstage.Client
}

type apiDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *apiSpecModel         `tfsdk:"spec"`
}

type apiSpecModel struct {
	Type       types.String `tfsdk:"type"`
	Lifecycle  types.String `tfsdk:"lifecycle"`
	Owner      types.String `tfsdk:"owner"`
	Definition types.String `tfsdk:"definition"`
	System     types.String `tfsdk:"system"`
}

// Metadata returns the data source type name.
func (d *apiDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api"
}

const (
	descriptionApiSpecType       = "Type of the API definition."
	descriptionApiSpecLifecycle  = "Lifecycle state of the API."
	descriptionApiSpecOwner      = "An entity reference to the owner of the API"
	descriptionApiSpecDefinition = "Definition of the API, based on the format defined by the type."
	descriptionApiSpecSystem     = "An entity reference to the system that the API belongs to."
)

// Schema defines the schema for the data source.
func (d *apiDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get a specific " +
			"[API entity](https://backstage.io/docs/features/software-catalog/descriptor-format#kind-api) from Backstage Software Catalog.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataUID},
			"name":        schema.StringAttribute{Required: true, Description: descriptionEntityMetadataName},
			"namespace":   schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataNamespace},
			"api_version": schema.StringAttribute{Computed: true, Description: descriptionEntityApiVersion},
			"kind":        schema.StringAttribute{Computed: true, Description: descriptionEntityKind},
			"metadata": schema.SingleNestedAttribute{Computed: true, Description: descriptionEntityMetadata, Attributes: map[string]schema.Attribute{
				"uid":         schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataUID},
				"etag":        schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataEtag},
				"name":        schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataName},
				"namespace":   schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataNamespace},
				"title":       schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataTitle},
				"description": schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataDescription},
				"labels":      schema.MapAttribute{Computed: true, Description: descriptionEntityMetadataLabels, ElementType: types.StringType},
				"annotations": schema.MapAttribute{Computed: true, Description: descriptionEntityMetadataAnnotations, ElementType: types.StringType},
				"tags":        schema.ListAttribute{Computed: true, Description: descriptionEntityMetadataTags, ElementType: types.StringType},
				"links": schema.ListNestedAttribute{Computed: true, Description: descriptionEntityMetadataLinks, NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"url":   schema.StringAttribute{Computed: true, Description: descriptionEntityLinkURL},
						"title": schema.StringAttribute{Computed: true, Description: descriptionEntityLinkTitle},
						"icon":  schema.StringAttribute{Computed: true, Description: descriptionEntityLinkIco},
						"type":  schema.StringAttribute{Computed: true, Description: descriptionEntityLinkType},
					},
				}},
			}},
			"relations": schema.ListNestedAttribute{Computed: true, Description: descriptionEntityRelations, NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type":       schema.StringAttribute{Computed: true, Description: descriptionEntityRelationType},
					"target_ref": schema.StringAttribute{Computed: true, Description: descriptionEntityRelationTargetRef},
					"target": schema.SingleNestedAttribute{Computed: true, Description: descriptionEntityRelationTarget,
						Attributes: map[string]schema.Attribute{
							"name":      schema.StringAttribute{Computed: true, Description: descriptionEntityRelationTargetName},
							"kind":      schema.StringAttribute{Computed: true, Description: descriptionEntityRelationTargetKind},
							"namespace": schema.StringAttribute{Computed: true, Description: descriptionEntityRelationTargetNamespace},
						}},
				},
			}},
			"spec": schema.SingleNestedAttribute{Computed: true, Description: descriptionEntitySpec, Attributes: map[string]schema.Attribute{
				"type":       schema.StringAttribute{Computed: true, Description: descriptionApiSpecType},
				"lifecycle":  schema.StringAttribute{Computed: true, Description: descriptionApiSpecLifecycle},
				"owner":      schema.StringAttribute{Computed: true, Description: descriptionApiSpecOwner},
				"definition": schema.StringAttribute{Computed: true, Description: descriptionApiSpecDefinition},
				"system":     schema.StringAttribute{Computed: true, Description: descriptionApiSpecSystem},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *apiDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *apiDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apiDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting API kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	api, response, err := d.client.Catalog.APIs.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Backstage API kind",
			fmt.Sprintf("Could not read Backstage API kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error reading Backstage API kind",
			fmt.Sprintf("Could not read Backstage API kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status),
		)
		return
	}

	state.ID = types.StringValue(api.Metadata.UID)
	state.ApiVersion = types.StringValue(api.ApiVersion)
	state.Kind = types.StringValue(api.Kind)

	for _, i := range api.Relations {
		state.Relations = append(state.Relations, entityRelationModel{
			Type:      types.StringValue(i.Type),
			TargetRef: types.StringValue(i.TargetRef),
			Target: &entityRelationTargetModel{
				Kind:      types.StringValue(i.Target.Kind),
				Name:      types.StringValue(i.Target.Name),
				Namespace: types.StringValue(i.Target.Namespace)},
		})
	}

	state.Spec = &apiSpecModel{
		Type:       types.StringValue(api.Spec.Type),
		Lifecycle:  types.StringValue(api.Spec.Lifecycle),
		Owner:      types.StringValue(api.Spec.Owner),
		Definition: types.StringValue(api.Spec.Definition),
		System:     types.StringValue(api.Spec.System),
	}

	state.Metadata = &entityMetadataModel{
		UID:         types.StringValue(api.Metadata.UID),
		Etag:        types.StringValue(api.Metadata.Etag),
		Name:        types.StringValue(api.Metadata.Name),
		Namespace:   types.StringValue(api.Metadata.Namespace),
		Title:       types.StringValue(api.Metadata.Title),
		Description: types.StringValue(api.Metadata.Description),
		Annotations: map[string]string{},
		Labels:      map[string]string{},
	}

	for k, v := range api.Metadata.Labels {
		state.Metadata.Labels[k] = v
	}

	for k, v := range api.Metadata.Annotations {
		state.Metadata.Annotations[k] = v
	}

	for _, v := range api.Metadata.Tags {
		state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
	}

	for _, v := range api.Metadata.Links {
		state.Metadata.Links = append(state.Metadata.Links, entityLinkModel{
			URL:   types.StringValue(v.URL),
			Title: types.StringValue(v.Title),
			Icon:  types.StringValue(v.Icon),
			Type:  types.StringValue(v.Type),
		})
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
