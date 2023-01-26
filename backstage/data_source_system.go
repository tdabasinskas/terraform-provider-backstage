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
	_ datasource.DataSource              = &systemDataSource{}
	_ datasource.DataSourceWithConfigure = &systemDataSource{}
)

// NewSystemDataSource is a helper function to simplify the provider implementation.
func NewSystemDataSource() datasource.DataSource {
	return &systemDataSource{}
}

// systemDataSource is the data source implementation.
type systemDataSource struct {
	client *backstage.Client
}

type systemDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *systemSpecModel      `tfsdk:"spec"`
}

type systemSpecModel struct {
	Owner  types.String `tfsdk:"owner"`
	Domain types.String `tfsdk:"domain"`
}

const (
	descriptionSystemSpecOwner  = "An entity reference to the owner of the system."
	descriptionSystemSpecDomain = "An entity reference to the domain that the system belongs to."
)

// Metadata returns the data source type name.
func (d *systemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

// Schema defines the schema for the data source.
func (d *systemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				"owner":  schema.StringAttribute{Computed: true, Description: descriptionSystemSpecOwner},
				"domain": schema.StringAttribute{Computed: true, Description: descriptionSystemSpecDomain},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *systemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *systemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state systemDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting System kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	system, response, err := d.client.Catalog.Systems.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Backstage System kind",
			fmt.Sprintf("Could not read Backstage System kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error reading Backstage System kind",
			fmt.Sprintf("Could not read Backstage System kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status),
		)
		return
	}

	state.ID = types.StringValue(system.Metadata.UID)
	state.ApiVersion = types.StringValue(system.ApiVersion)
	state.Kind = types.StringValue(system.Kind)

	for _, i := range system.Relations {
		state.Relations = append(state.Relations, entityRelationModel{
			Type:      types.StringValue(i.Type),
			TargetRef: types.StringValue(i.TargetRef),
			Target: &entityRelationTargetModel{
				Kind:      types.StringValue(i.Target.Kind),
				Name:      types.StringValue(i.Target.Name),
				Namespace: types.StringValue(i.Target.Namespace)},
		})
	}

	state.Spec = &systemSpecModel{
		Owner:  types.StringValue(system.Spec.Owner),
		Domain: types.StringValue(system.Spec.Domain),
	}

	state.Metadata = &entityMetadataModel{
		UID:         types.StringValue(system.Metadata.UID),
		Etag:        types.StringValue(system.Metadata.Etag),
		Name:        types.StringValue(system.Metadata.Name),
		Namespace:   types.StringValue(system.Metadata.Namespace),
		Title:       types.StringValue(system.Metadata.Title),
		Description: types.StringValue(system.Metadata.Description),
		Annotations: map[string]string{},
		Labels:      map[string]string{},
	}

	for k, v := range system.Metadata.Labels {
		state.Metadata.Labels[k] = v
	}

	for k, v := range system.Metadata.Annotations {
		state.Metadata.Annotations[k] = v
	}

	for _, v := range system.Metadata.Tags {
		state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
	}

	for _, v := range system.Metadata.Links {
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
