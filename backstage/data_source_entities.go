package backstage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/datolabs-io/go-backstage/v3"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &entityDataSource{}
	_ datasource.DataSourceWithConfigure = &entityDataSource{}
)

// NewEntityDataSource is a helper function to simplify the provider implementation.
func NewEntityDataSource() datasource.DataSource {
	return &entityDataSource{}
}

// entityDataSource is the data source implementation.
type entityDataSource struct {
	client *backstage.Client
}

type entityDataSourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Filters  []string             `tfsdk:"filters"`
	Entities []entityModel        `tfsdk:"entities"`
	Fallback *entityFallbackModel `tfsdk:"fallback"`
}

type entityModel struct {
	ApiVersion types.String          `tfsdk:"api_version"`
	Spec       jsontypes.Normalized  `tfsdk:"spec"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
}

type entityMetadataModel struct {
	UID         types.String      `tfsdk:"uid"`
	Etag        types.String      `tfsdk:"etag"`
	Name        types.String      `tfsdk:"name"`
	Namespace   types.String      `tfsdk:"namespace"`
	Title       types.String      `tfsdk:"title"`
	Description types.String      `tfsdk:"description"`
	Annotations map[string]string `tfsdk:"annotations"`
	Labels      map[string]string `tfsdk:"labels"`
	Tags        []types.String    `tfsdk:"tags"`
	Links       []entityLinkModel `tfsdk:"links"`
}

type entityRelationModel struct {
	Type      types.String               `tfsdk:"type"`
	TargetRef types.String               `tfsdk:"target_ref"`
	Target    *entityRelationTargetModel `tfsdk:"target"`
}

type entityRelationTargetModel struct {
	Name      types.String `tfsdk:"name"`
	Kind      types.String `tfsdk:"kind"`
	Namespace types.String `tfsdk:"namespace"`
}

type entityLinkModel struct {
	URL   types.String `tfsdk:"url"`
	Title types.String `tfsdk:"title"`
	Icon  types.String `tfsdk:"icon"`
	Type  types.String `tfsdk:"type"`
}

type entityFallbackModel struct {
	ID       types.String  `tfsdk:"id"`
	Filters  []string      `tfsdk:"filters"`
	Entities []entityModel `tfsdk:"entities"`
}

const (
	patternEntityName                  = `^[a-zA-Z0-9\-_\.]*$`
	descriptionEntityFilters           = "A set of conditions that can be used to filter entities."
	descriptionEntitySpec              = "The specification data describing the entity itself."
	descriptionEntitySpecJson          = "The specification data describing the entity itself (as JSON)."
	descriptionEntityApiVersion        = "Version of specification format for this particular entity that this is written against."
	descriptionEntityKind              = "The high level entity type being described."
	descriptionEntityMetadata          = "Metadata fields common to all versions/kinds of entity."
	descriptionEntityMetadataName      = "Name of the entity."
	descriptionEntityMetadataNamespace = "Namespace that the entity belongs to."
	descriptionEntityMetadataUID       = "A globally unique ID for the entity. This field can not be set by the user at creation time, and the server will reject an " +
		"attempt to do so. The field will be populated in read operations."
	descriptionEntityMetadataEtag = "An opaque string that changes for each update operation to any part of the entity, including metadata. This field can not be " +
		"set by the user at creation time, and the server will reject an attempt to do so. The field will be populated in read operations.The field can (optionally) be " +
		"specified when performing update or delete operations, and the server will then reject the operation if it does not match the current stored value."
	descriptionEntityMetadataTitle       = "A display name of the entity, to be presented in user interfaces instead of the name property, when available."
	descriptionEntityMetadataDescription = "A short (typically relatively few words) description of the entity."
	descriptionEntityMetadataLabels      = "Key/Value pairs of identifying information attached to the entity."
	descriptionEntityMetadataAnnotations = "Key/Value pairs of non-identifying auxiliary information attached to entity."
	descriptionEntityMetadataTags        = "A list of single-valued strings, to for example classify catalog entities in various ways."
	descriptionEntityMetadataLinks       = "A list of external hyperlinks related to the entity. Links can provide additional contextual information that may be " +
		"located outside of Backstage itself. For example, an admin dashboard or external CMS page."
	descriptionEntityLinkURL                 = "URL in a standard uri format."
	descriptionEntityLinkTitle               = "A user-friendly display name for the link."
	descriptionEntityLinkIco                 = "A key representing a visual icon to be displayed in the UI."
	descriptionEntityLinkType                = "An optional value to categorize links into specific groups."
	descriptionEntityRelations               = "Relations that this entity has with other entities"
	descriptionEntityRelationType            = "Type of the relation."
	descriptionEntityRelationTargetRef       = "The entity ref of the target of this relation."
	descriptionEntityRelationTarget          = "The entity of the target of this relation."
	descriptionEntityRelationTargetName      = "Name of the entity."
	descriptionEntityRelationTargetKind      = "The high level entity type being described."
	descriptionEntityRelationTargetNamespace = "Namespace that the target entity belongs to."
	descriptionEntityFallback                = "A complete replica of the `Entity` as it would exist in backstage. Set this to provide a fallback in case the Backstage instance is not functioning, is down, or is unrealiable."
)

// Metadata returns the data source type name.
func (d *entityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entities"
}

// Schema defines the schema for the data source.
func (d *entityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get a filtered list of " +
			"[entities](https://backstage.io/docs/features/software-catalog/descriptor-format#overall-shape-of-an-entity) from Backstage Software Catalog. For more " +
			"information about the way filters are defined and applied, see " +
			"[Backstage documentation](https://backstage.io/docs/features/software-catalog/software-catalog-api#filtering).",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, Description: descriptionEntityMetadataUID},
			"filters": schema.ListAttribute{Required: true, Description: descriptionEntityFilters, ElementType: types.StringType},
			"entities": schema.ListNestedAttribute{Computed: true, Description: descriptionEntitySpec, NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"api_version": schema.StringAttribute{Computed: true, Description: descriptionEntityApiVersion},
					"spec":        schema.StringAttribute{Computed: true, Description: descriptionEntitySpecJson, CustomType: jsontypes.NormalizedType{}},
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
				},
			}},
			"fallback": schema.SingleNestedAttribute{Optional: true, Description: descriptionEntityFallback, Attributes: map[string]schema.Attribute{
				"id":      schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataUID},
				"filters": schema.ListAttribute{Required: true, Description: descriptionEntityFilters, ElementType: types.StringType},
				"entities": schema.ListNestedAttribute{Optional: true, Description: descriptionEntitySpec, NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_version": schema.StringAttribute{Optional: true, Description: descriptionEntityApiVersion},
						"spec":        schema.StringAttribute{Optional: true, Description: descriptionEntitySpecJson, CustomType: jsontypes.NormalizedType{}},
						"kind":        schema.StringAttribute{Optional: true, Description: descriptionEntityKind},
						"metadata": schema.SingleNestedAttribute{Optional: true, Description: descriptionEntityMetadata, Attributes: map[string]schema.Attribute{
							"uid":         schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataUID},
							"etag":        schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataEtag},
							"name":        schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataName},
							"namespace":   schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataNamespace},
							"title":       schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataTitle},
							"description": schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataDescription},
							"labels":      schema.MapAttribute{Optional: true, Description: descriptionEntityMetadataLabels, ElementType: types.StringType},
							"annotations": schema.MapAttribute{Optional: true, Description: descriptionEntityMetadataAnnotations, ElementType: types.StringType},
							"tags":        schema.ListAttribute{Optional: true, Description: descriptionEntityMetadataTags, ElementType: types.StringType},
							"links": schema.ListNestedAttribute{Optional: true, Description: descriptionEntityMetadataLinks, NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"url":   schema.StringAttribute{Optional: true, Description: descriptionEntityLinkURL},
									"title": schema.StringAttribute{Optional: true, Description: descriptionEntityLinkTitle},
									"icon":  schema.StringAttribute{Optional: true, Description: descriptionEntityLinkIco},
									"type":  schema.StringAttribute{Optional: true, Description: descriptionEntityLinkType},
								},
							}},
						}},
						"relations": schema.ListNestedAttribute{Optional: true, Description: descriptionEntityRelations, NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type":       schema.StringAttribute{Optional: true, Description: descriptionEntityRelationType},
								"target_ref": schema.StringAttribute{Optional: true, Description: descriptionEntityRelationTargetRef},
								"target": schema.SingleNestedAttribute{Optional: true, Description: descriptionEntityRelationTarget,
									Attributes: map[string]schema.Attribute{
										"name":      schema.StringAttribute{Optional: true, Description: descriptionEntityRelationTargetName},
										"kind":      schema.StringAttribute{Optional: true, Description: descriptionEntityRelationTargetKind},
										"namespace": schema.StringAttribute{Optional: true, Description: descriptionEntityRelationTargetNamespace},
									}},
							},
						}},
					},
				}},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *entityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *entityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state entityDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting entities %v from Backstage API", state.Filters))
	entities, response, err := d.client.Catalog.Entities.List(ctx, &backstage.ListEntityOptions{
		Filters: state.Filters,
		Order:   []backstage.ListEntityOrder{{Field: "metadata.name", Direction: "asc"}},
	})
	if err != nil {
		const shortErr = "Error reading Backstage entities"
		longErr := fmt.Sprintf("Could not read Backstage entities %v: %s", state.Filters, err.Error())
		if state.Fallback == nil {
			resp.Diagnostics.AddError(shortErr, longErr)
			return
		}
		resp.Diagnostics.AddWarning(shortErr, longErr)
	}

	if response.StatusCode != http.StatusOK {
		const shortErr = "Error reading Backstage entities"
		longErr := fmt.Sprintf("Could not read Backstage entities %v: %s", state.Filters, response.Status)
		if state.Fallback == nil {
			resp.Diagnostics.AddError(shortErr, longErr)
			return
		}
		resp.Diagnostics.AddWarning(shortErr, longErr)
	}
	if (err != nil || response.StatusCode != http.StatusOK) && state.Fallback != nil {
		if state.Fallback.ID.IsNull() {
			state.Fallback.ID = types.StringValue("123456789")
		}
		state.ID = state.Fallback.ID
		state.Filters = state.Fallback.Filters
		state.Entities = state.Fallback.Entities
	}

	if err == nil && response.StatusCode == http.StatusOK {
		state.ID = types.StringValue(fmt.Sprint(state.Filters))

		for _, e := range entities {
			v, err := json.Marshal(e.Spec)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error parsing Backstage entity specs",
					fmt.Sprintf("Could not parse Specs for Backstage entity %v: %s", e.Metadata.Name, err.Error()),
				)
				continue
			}

			entity := entityModel{
				ApiVersion: types.StringValue(e.ApiVersion),
				Kind:       types.StringValue(e.Kind),
				Spec:       jsontypes.NewNormalizedValue(string(v)),
			}

			for _, i := range e.Relations {
				entity.Relations = append(entity.Relations, entityRelationModel{
					Type:      types.StringValue(i.Type),
					TargetRef: types.StringValue(i.TargetRef),
					Target: &entityRelationTargetModel{
						Kind:      types.StringValue(i.Target.Kind),
						Name:      types.StringValue(i.Target.Name),
						Namespace: types.StringValue(i.Target.Namespace)},
				})
			}

			entity.Metadata = &entityMetadataModel{
				UID:         types.StringValue(e.Metadata.UID),
				Etag:        types.StringValue(e.Metadata.Etag),
				Name:        types.StringValue(e.Metadata.Name),
				Namespace:   types.StringValue(e.Metadata.Namespace),
				Title:       types.StringValue(e.Metadata.Title),
				Description: types.StringValue(e.Metadata.Description),
				Annotations: map[string]string{},
				Labels:      map[string]string{},
			}

			for k, v := range e.Metadata.Labels {
				entity.Metadata.Labels[k] = v
			}

			for k, v := range e.Metadata.Annotations {
				entity.Metadata.Annotations[k] = v
			}

			for _, v := range e.Metadata.Tags {
				entity.Metadata.Tags = append(entity.Metadata.Tags, types.StringValue(v))
			}

			for _, v := range e.Metadata.Links {
				entity.Metadata.Links = append(entity.Metadata.Links, entityLinkModel{
					URL:   types.StringValue(v.URL),
					Title: types.StringValue(v.Title),
					Icon:  types.StringValue(v.Icon),
					Type:  types.StringValue(v.Type),
				})
			}

			state.Entities = append(state.Entities, entity)
		}
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
