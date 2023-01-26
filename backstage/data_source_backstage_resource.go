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
	_ datasource.DataSource              = &resourceDataSource{}
	_ datasource.DataSourceWithConfigure = &resourceDataSource{}
)

// NewResourceDataSource is a helper function to simplify the provider implementation.
func NewResourceDataSource() datasource.DataSource {
	return &resourceDataSource{}
}

// resourceDataSource is the data source implementation.
type resourceDataSource struct {
	client *backstage.Client
}

type resourceDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *resourceSpecModel    `tfsdk:"spec"`
}

type resourceSpecModel struct {
	Type      types.String   `tfsdk:"type"`
	Owner     types.String   `tfsdk:"owner"`
	DependsOn []types.String `tfsdk:"depends_on"`
	System    types.String   `tfsdk:"system"`
}

const (
	descriptionResourceSpecType      = "Type of the resource definition."
	descriptionResourceSpecOwner     = "An entity reference to the owner of the resource"
	descriptionResourceSpecDependsOn = "An array of references to other entities that the resource depends on to function."
	descriptionResourceSpecSystem    = "An entity reference to the system that the resource belongs to."
)

// Metadata returns the data source type name.
func (d *resourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

// Schema defines the schema for the data source.
func (d *resourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				"type":       schema.StringAttribute{Computed: true, Description: descriptionResourceSpecType},
				"owner":      schema.StringAttribute{Computed: true, Description: descriptionResourceSpecOwner},
				"depends_on": schema.ListAttribute{Computed: true, Description: descriptionResourceSpecDependsOn, ElementType: types.StringType},
				"system":     schema.StringAttribute{Computed: true, Description: descriptionResourceSpecSystem},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *resourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *resourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state resourceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting Resource kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	resource, response, err := d.client.Catalog.Resources.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Backstage Resource kind",
			fmt.Sprintf("Could not read Backstage Resource kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error reading Backstage Resource kind",
			fmt.Sprintf("Could not read Backstage Resource kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status),
		)
		return
	}

	state.ID = types.StringValue(resource.Metadata.UID)
	state.ApiVersion = types.StringValue(resource.ApiVersion)
	state.Kind = types.StringValue(resource.Kind)

	for _, i := range resource.Relations {
		state.Relations = append(state.Relations, entityRelationModel{
			Type:      types.StringValue(i.Type),
			TargetRef: types.StringValue(i.TargetRef),
			Target: &entityRelationTargetModel{
				Kind:      types.StringValue(i.Target.Kind),
				Name:      types.StringValue(i.Target.Name),
				Namespace: types.StringValue(i.Target.Namespace)},
		})
	}

	state.Spec = &resourceSpecModel{
		Type:   types.StringValue(resource.Spec.Type),
		Owner:  types.StringValue(resource.Spec.Owner),
		System: types.StringValue(resource.Spec.System),
	}

	for _, i := range resource.Spec.DependsOn {
		state.Spec.DependsOn = append(state.Spec.DependsOn, types.StringValue(i))
	}

	state.Metadata = &entityMetadataModel{
		UID:         types.StringValue(resource.Metadata.UID),
		Etag:        types.StringValue(resource.Metadata.Etag),
		Name:        types.StringValue(resource.Metadata.Name),
		Namespace:   types.StringValue(resource.Metadata.Namespace),
		Title:       types.StringValue(resource.Metadata.Title),
		Description: types.StringValue(resource.Metadata.Description),
		Annotations: map[string]string{},
		Labels:      map[string]string{},
	}

	for k, v := range resource.Metadata.Labels {
		state.Metadata.Labels[k] = v
	}

	for k, v := range resource.Metadata.Annotations {
		state.Metadata.Annotations[k] = v
	}

	for _, v := range resource.Metadata.Tags {
		state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
	}

	for _, v := range resource.Metadata.Links {
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
