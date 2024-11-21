package backstage

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/datolabs-io/go-backstage/v3"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &componentDataSource{}
	_ datasource.DataSourceWithConfigure = &componentDataSource{}
)

// NewComponentDataSource is a helper function to simplify the provider implementation.
func NewComponentDataSource() datasource.DataSource {
	return &componentDataSource{}
}

// componentDataSource is the data source implementation.
type componentDataSource struct {
	client *backstage.Client
}

type componentDataSourceModel struct {
	ID         types.String            `tfsdk:"id"`
	Name       types.String            `tfsdk:"name"`
	Namespace  types.String            `tfsdk:"namespace"`
	ApiVersion types.String            `tfsdk:"api_version"`
	Kind       types.String            `tfsdk:"kind"`
	Metadata   *entityMetadataModel    `tfsdk:"metadata"`
	Relations  []entityRelationModel   `tfsdk:"relations"`
	Spec       *componentSpecModel     `tfsdk:"spec"`
	Fallback   *componentFallbackModel `tfsdk:"fallback"`
}

type componentSpecModel struct {
	Type           types.String   `tfsdk:"type"`
	Lifecycle      types.String   `tfsdk:"lifecycle"`
	Owner          types.String   `tfsdk:"owner"`
	SubcomponentOf types.String   `tfsdk:"subcomponent_of"`
	ProvidesApis   []types.String `tfsdk:"provides_apis"`
	ConsumesApis   []types.String `tfsdk:"consumes_apis"`
	DependsOn      []types.String `tfsdk:"depends_on"`
	System         types.String   `tfsdk:"system"`
}

type componentFallbackModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *componentSpecModel   `tfsdk:"spec"`
}

const (
	descriptionComponentSpecType           = "Type of the component definition."
	descriptionComponentSpecLifecycle      = "Lifecycle state of the component."
	descriptionComponentSpecOwner          = "An entity reference to the owner of the component"
	descriptionComponentSpecSubcomponentOf = "An entity reference to another component of which the component is a part."
	descriptionComponentSpecProvidesAPIs   = "An array of entity references to the APIs that are provided by the component."
	descriptionComponentSpecConsumesAPIs   = "An array of entity references to the APIs that are consumed by the component."
	descriptionComponentSpecDependsOn      = "An array of entity references to the components and resources that the component depends on."
	descriptionComponentSpecSystem         = "An entity reference to the system that the component belongs to."
	descriptionComponentFallback           = "A full entity object that represents the Component as it would exist in backstage. Set this to provide a fallback in case the Component is not functioning, is down, or is unrealiable."
)

// Metadata returns the data source type name.
func (d *componentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

// Schema defines the schema for the data source.
func (d *componentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get a specific " +
			"[Component entity](https://backstage.io/docs/features/software-catalog/descriptor-format#kind-component) from Backstage Software Catalog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataUID},
			"name": schema.StringAttribute{Required: true, Description: descriptionEntityMetadataName, Validators: []validator.String{
				stringvalidator.LengthBetween(1, 63),
				stringvalidator.RegexMatches(
					regexp.MustCompile(patternEntityName),
					"must follow Backstage format restrictions",
				),
			}},
			"namespace": schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataNamespace, Validators: []validator.String{
				stringvalidator.LengthBetween(1, 63),
				stringvalidator.RegexMatches(
					regexp.MustCompile(patternEntityName),
					"must follow Backstage format restrictions",
				),
			}},
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
				"type":            schema.StringAttribute{Computed: true, Description: descriptionComponentSpecType},
				"lifecycle":       schema.StringAttribute{Computed: true, Description: descriptionComponentSpecLifecycle},
				"owner":           schema.StringAttribute{Computed: true, Description: descriptionComponentSpecOwner},
				"subcomponent_of": schema.StringAttribute{Computed: true, Description: descriptionComponentSpecSubcomponentOf},
				"provides_apis":   schema.ListAttribute{Computed: true, Description: descriptionComponentSpecProvidesAPIs, ElementType: types.StringType},
				"consumes_apis":   schema.ListAttribute{Computed: true, Description: descriptionComponentSpecConsumesAPIs, ElementType: types.StringType},
				"depends_on":      schema.ListAttribute{Computed: true, Description: descriptionComponentSpecDependsOn, ElementType: types.StringType},
				"system":          schema.StringAttribute{Computed: true, Description: descriptionComponentSpecSystem},
			}},
			"fallback": schema.SingleNestedAttribute{Computed: true, Optional: true, Description: descriptionComponentFallback, Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataUID},
				"name": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataName, Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(patternEntityName),
						"must follow Backstage format restrictions",
					),
				}},
				"namespace": schema.StringAttribute{Optional: true, Computed: true, Description: descriptionEntityMetadataNamespace, Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(patternEntityName),
						"must follow Backstage format restrictions",
					),
				}},
				"api_version": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityApiVersion},
				"kind":        schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityKind},
				"metadata": schema.SingleNestedAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadata, Attributes: map[string]schema.Attribute{
					"uid":         schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataUID},
					"etag":        schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataEtag},
					"name":        schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataName},
					"namespace":   schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataNamespace},
					"title":       schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataTitle},
					"description": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataDescription},
					"labels":      schema.MapAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataLabels, ElementType: types.StringType},
					"annotations": schema.MapAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataAnnotations, ElementType: types.StringType},
					"tags":        schema.ListAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataTags, ElementType: types.StringType},
					"links": schema.ListNestedAttribute{Computed: true, Optional: true, Description: descriptionEntityMetadataLinks, NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"url":   schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityLinkURL},
							"title": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityLinkTitle},
							"icon":  schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityLinkIco},
							"type":  schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityLinkType},
						},
					}},
				}},
				"relations": schema.ListNestedAttribute{Computed: true, Optional: true, Description: descriptionEntityRelations, NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":       schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityRelationType},
						"target_ref": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityRelationTargetRef},
						"target": schema.SingleNestedAttribute{Computed: true, Optional: true, Description: descriptionEntityRelationTarget,
							Attributes: map[string]schema.Attribute{
								"name":      schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityRelationTargetName},
								"kind":      schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityRelationTargetKind},
								"namespace": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionEntityRelationTargetNamespace},
							}},
					},
				}},
				"spec": schema.SingleNestedAttribute{Computed: true, Optional: true, Description: descriptionEntitySpec, Attributes: map[string]schema.Attribute{
					"type":            schema.StringAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecType},
					"lifecycle":       schema.StringAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecLifecycle},
					"owner":           schema.StringAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecOwner},
					"subcomponent_of": schema.StringAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecSubcomponentOf},
					"provides_apis":   schema.ListAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecProvidesAPIs, ElementType: types.StringType},
					"consumes_apis":   schema.ListAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecConsumesAPIs, ElementType: types.StringType},
					"depends_on":      schema.ListAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecDependsOn, ElementType: types.StringType},
					"system":          schema.StringAttribute{Computed: true, Optional: true, Description: descriptionComponentSpecSystem},
				}},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *componentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *componentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state componentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting Component kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	component, response, err := d.client.Catalog.Components.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		const shortErr = "Error reading Backstage Component kind"
		description := fmt.Sprintf("Could not read Backstage Component kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error())
		if state.Fallback == nil {
			resp.Diagnostics.AddError(shortErr, description)
			return
		}
		resp.Diagnostics.AddWarning(shortErr, description)
	}

	if response.StatusCode != http.StatusOK {
		const shortErr = "Error reading Backstage Component kind"
		description := fmt.Sprintf("Could not read Backstage Component kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status)
		if state.Fallback == nil {
			resp.Diagnostics.AddError(shortErr, description)
			return
		}
		resp.Diagnostics.AddWarning(shortErr, description)
	}
	if (err != nil || response.StatusCode != http.StatusOK) && state.Fallback != nil {
		state.ID = state.Fallback.ID
		state.Name = state.Fallback.Name
		state.Namespace = state.Fallback.Namespace
		state.ApiVersion = state.Fallback.ApiVersion
		state.Kind = state.Fallback.Kind
		state.Metadata = state.Fallback.Metadata
		state.Relations = state.Fallback.Relations
		state.Spec = state.Fallback.Spec
	}

	if err == nil && response.StatusCode == http.StatusOK {
		state.ID = types.StringValue(component.Metadata.UID)
		state.ApiVersion = types.StringValue(component.ApiVersion)
		state.Kind = types.StringValue(component.Kind)

		for _, i := range component.Relations {
			state.Relations = append(state.Relations, entityRelationModel{
				Type:      types.StringValue(i.Type),
				TargetRef: types.StringValue(i.TargetRef),
				Target: &entityRelationTargetModel{
					Kind:      types.StringValue(i.Target.Kind),
					Name:      types.StringValue(i.Target.Name),
					Namespace: types.StringValue(i.Target.Namespace)},
			})
		}

		state.Spec = &componentSpecModel{
			Type:           types.StringValue(component.Spec.Type),
			Lifecycle:      types.StringValue(component.Spec.Lifecycle),
			Owner:          types.StringValue(component.Spec.Owner),
			SubcomponentOf: types.StringValue(component.Spec.SubcomponentOf),
			System:         types.StringValue(component.Spec.System),
		}

		for _, i := range component.Spec.ProvidesApis {
			state.Spec.ProvidesApis = append(state.Spec.ProvidesApis, types.StringValue(i))
		}

		for _, i := range component.Spec.ConsumesApis {
			state.Spec.ConsumesApis = append(state.Spec.ConsumesApis, types.StringValue(i))
		}

		for _, i := range component.Spec.DependsOn {
			state.Spec.DependsOn = append(state.Spec.DependsOn, types.StringValue(i))
		}

		state.Metadata = &entityMetadataModel{
			UID:         types.StringValue(component.Metadata.UID),
			Etag:        types.StringValue(component.Metadata.Etag),
			Name:        types.StringValue(component.Metadata.Name),
			Namespace:   types.StringValue(component.Metadata.Namespace),
			Title:       types.StringValue(component.Metadata.Title),
			Description: types.StringValue(component.Metadata.Description),
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		}

		for k, v := range component.Metadata.Labels {
			state.Metadata.Labels[k] = v
		}

		for k, v := range component.Metadata.Annotations {
			state.Metadata.Annotations[k] = v
		}

		for _, v := range component.Metadata.Tags {
			state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
		}

		for _, v := range component.Metadata.Links {
			state.Metadata.Links = append(state.Metadata.Links, entityLinkModel{
				URL:   types.StringValue(v.URL),
				Title: types.StringValue(v.Title),
				Icon:  types.StringValue(v.Icon),
				Type:  types.StringValue(v.Type),
			})
		}
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
