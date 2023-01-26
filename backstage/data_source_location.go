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
	_ datasource.DataSource              = &locationDataSource{}
	_ datasource.DataSourceWithConfigure = &locationDataSource{}
)

// NewLocationDataSource is a helper function to simplify the provider implementation.
func NewLocationDataSource() datasource.DataSource {
	return &locationDataSource{}
}

// locationDataSource is the data source implementation.
type locationDataSource struct {
	client *backstage.Client
}

type locationDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *locationSpecModel    `tfsdk:"spec"`
}

type locationSpecModel struct {
	Type     types.String   `tfsdk:"type"`
	Target   types.String   `tfsdk:"target"`
	Targets  []types.String `tfsdk:"targets"`
	Presence types.String   `tfsdk:"presence"`
}

const (
	descriptionLocationSpecType = "The single location type, that's common to the targets specified in the spec. If it is left out, it is inherited from the location type " +
		"that originally read the entity data."
	descriptionLocationSpecTarget = "Target as a string. Can be either an absolute path/URL (depending on the type), or a relative path such as./details/catalog-info.yaml " +
		"which is resolved relative to the location of this Location entity itself."
	descriptionLocationSpecTargets = "A list of targets as strings. They can all be either absolute paths/URLs (depending on the type), or relative paths such as" +
		"./details/catalog-info.yaml which are resolved relative to the location of this Location entity itself."
	descriptionLocationSpecPresence = "Presence describes whether the presence of the location target is required and it should be considered an error if it can not be found."
)

// Metadata returns the data source type name.
func (d *locationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

// Schema defines the schema for the data source.
func (d *locationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				"type":     schema.StringAttribute{Computed: true, Description: descriptionLocationSpecType},
				"target":   schema.StringAttribute{Computed: true, Description: descriptionLocationSpecTarget},
				"targets":  schema.ListAttribute{Computed: true, Description: descriptionLocationSpecTargets, ElementType: types.StringType},
				"presence": schema.StringAttribute{Computed: true, Description: descriptionLocationSpecPresence},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *locationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *locationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state locationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting Location kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	location, response, err := d.client.Catalog.Locations.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Backstage Location kind",
			fmt.Sprintf("Could not read Backstage Location kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error reading Backstage Location kind",
			fmt.Sprintf("Could not read Backstage Location kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status),
		)
		return
	}

	state.ID = types.StringValue(location.Metadata.UID)
	state.ApiVersion = types.StringValue(location.ApiVersion)
	state.Kind = types.StringValue(location.Kind)

	for _, i := range location.Relations {
		state.Relations = append(state.Relations, entityRelationModel{
			Type:      types.StringValue(i.Type),
			TargetRef: types.StringValue(i.TargetRef),
			Target: &entityRelationTargetModel{
				Kind:      types.StringValue(i.Target.Kind),
				Name:      types.StringValue(i.Target.Name),
				Namespace: types.StringValue(i.Target.Namespace)},
		})
	}

	state.Spec = &locationSpecModel{
		Type:     types.StringValue(location.Spec.Type),
		Target:   types.StringValue(location.Spec.Target),
		Presence: types.StringValue(location.Spec.Presence),
	}

	for _, i := range location.Spec.Targets {
		state.Spec.Targets = append(state.Spec.Targets, types.StringValue(i))
	}

	state.Metadata = &entityMetadataModel{
		UID:         types.StringValue(location.Metadata.UID),
		Etag:        types.StringValue(location.Metadata.Etag),
		Name:        types.StringValue(location.Metadata.Name),
		Namespace:   types.StringValue(location.Metadata.Namespace),
		Title:       types.StringValue(location.Metadata.Title),
		Description: types.StringValue(location.Metadata.Description),
		Annotations: map[string]string{},
		Labels:      map[string]string{},
	}

	for k, v := range location.Metadata.Labels {
		state.Metadata.Labels[k] = v
	}

	for k, v := range location.Metadata.Annotations {
		state.Metadata.Annotations[k] = v
	}

	for _, v := range location.Metadata.Tags {
		state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
	}

	for _, v := range location.Metadata.Links {
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
