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
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

// NewGroupDataSource is a helper function to simplify the provider implementation.
func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

// groupDataSource is the data source implementation.
type groupDataSource struct {
	client *backstage.Client
}

type groupDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *groupSpecModel       `tfsdk:"spec"`
	Fallback   *groupFallbackModel   `tfsdk:"fallback"`
}

type groupSpecModel struct {
	Type     types.String           `tfsdk:"type"`
	Profile  *groupSpecProfileModel `tfsdk:"profile"`
	Parent   types.String           `tfsdk:"parent"`
	Children []types.String         `tfsdk:"children"`
	Members  []types.String         `tfsdk:"members"`
}

type groupSpecProfileModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	Email       types.String `tfsdk:"email"`
	Picture     types.String `tfsdk:"picture"`
}

type groupFallbackModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *groupSpecModel       `tfsdk:"spec"`
}

const (
	descriptionGroupType                   = "The type of group."
	descriptionGroupSpecProfile            = "Profile information about the group, mainly for display purposes."
	descriptionGroupSpecProfileDisplayName = "A simple display name to present to users."
	descriptionGroupSpecProfileEmail       = "Email where this entity can be reached."
	descriptionGroupSpecProfilePicture     = "A URL of an image that represents this entity."
	descriptionGroupSpecParent             = "Parent is the immediate parent group in the hierarchy, if any."
	descriptionGroupSpecChildren           = "Children contains immediate child groups of this group in the hierarchy (whose parent field points to this group)."
	descriptionGroupSpecMembers            = "Members contains the users that are members of this group."
	descriptionGroupFallback               = "A complete replica of the `Group` as it would exist in backstage. Set this to provide a fallback in case the Backstage instance is not functioning, is down, or is unrealiable."
)

// Metadata returns the data source type name.
func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the data source.
func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get a specific " +
			"[Group entity](https://backstage.io/docs/features/software-catalog/descriptor-format#kind-group) from Backstage Software Catalog.",
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
				"type":     schema.StringAttribute{Computed: true, Description: descriptionGroupType},
				"parent":   schema.StringAttribute{Computed: true, Description: descriptionGroupSpecParent},
				"children": schema.ListAttribute{Computed: true, Description: descriptionGroupSpecChildren, ElementType: types.StringType},
				"members":  schema.ListAttribute{Computed: true, Description: descriptionGroupSpecMembers, ElementType: types.StringType},
				"profile": schema.SingleNestedAttribute{Computed: true, Description: descriptionGroupSpecProfile, Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{Computed: true, Description: descriptionGroupSpecProfileDisplayName},
					"email":        schema.StringAttribute{Computed: true, Description: descriptionGroupSpecProfileEmail},
					"picture":      schema.StringAttribute{Computed: true, Description: descriptionGroupSpecProfilePicture},
				}},
			}},
			"fallback": schema.SingleNestedAttribute{Optional: true, Description: descriptionGroupFallback, Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Optional: true, Description: descriptionEntityMetadataUID},
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
				"api_version": schema.StringAttribute{Optional: true, Description: descriptionEntityApiVersion},
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
				"spec": schema.SingleNestedAttribute{Optional: true, Description: descriptionEntitySpec, Attributes: map[string]schema.Attribute{
					"type":     schema.StringAttribute{Optional: true, Description: descriptionGroupType},
					"parent":   schema.StringAttribute{Optional: true, Description: descriptionGroupSpecParent},
					"children": schema.ListAttribute{Optional: true, Description: descriptionGroupSpecChildren, ElementType: types.StringType},
					"members":  schema.ListAttribute{Optional: true, Description: descriptionGroupSpecMembers, ElementType: types.StringType},
					"profile": schema.SingleNestedAttribute{Optional: true, Description: descriptionGroupSpecProfile, Attributes: map[string]schema.Attribute{
						"display_name": schema.StringAttribute{Optional: true, Description: descriptionGroupSpecProfileDisplayName},
						"email":        schema.StringAttribute{Optional: true, Description: descriptionGroupSpecProfileEmail},
						"picture":      schema.StringAttribute{Optional: true, Description: descriptionGroupSpecProfilePicture},
					}},
				}},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting Group kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	group, response, err := d.client.Catalog.Groups.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		const shortErr = "Error reading Backstage Group kind"
		longErr := fmt.Sprintf("Could not read Backstage Group kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error())
		if state.Fallback == nil {
			resp.Diagnostics.AddError(shortErr, longErr)
			return
		}
		resp.Diagnostics.AddWarning(shortErr, longErr)
	}

	if response.StatusCode != http.StatusOK {
		const shortErr = "Error reading Backstage Group kind"
		longErr := fmt.Sprintf("Could not read Backstage Group kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status)
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
		if state.Fallback.ApiVersion.IsNull() {
			state.Fallback.ApiVersion = types.StringValue("backstage.io/v1alpha1")
		}
		if state.Fallback.Kind.IsNull() {
			state.Fallback.Kind = types.StringValue(backstage.KindGroup)
		}
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

		state.ID = types.StringValue(group.Metadata.UID)
		state.ApiVersion = types.StringValue(group.ApiVersion)
		state.Kind = types.StringValue(group.Kind)

		for _, i := range group.Relations {
			state.Relations = append(state.Relations, entityRelationModel{
				Type:      types.StringValue(i.Type),
				TargetRef: types.StringValue(i.TargetRef),
				Target: &entityRelationTargetModel{
					Kind:      types.StringValue(i.Target.Kind),
					Name:      types.StringValue(i.Target.Name),
					Namespace: types.StringValue(i.Target.Namespace)},
			})
		}

		state.Spec = &groupSpecModel{
			Type:   types.StringValue(group.Spec.Type),
			Parent: types.StringValue(group.Spec.Parent),
			Profile: &groupSpecProfileModel{
				DisplayName: types.StringValue(group.Spec.Profile.DisplayName),
				Email:       types.StringValue(group.Spec.Profile.Email),
				Picture:     types.StringValue(group.Spec.Profile.Picture),
			},
		}

		for _, i := range group.Spec.Children {
			state.Spec.Children = append(state.Spec.Children, types.StringValue(i))
		}

		for _, i := range group.Spec.Members {
			state.Spec.Members = append(state.Spec.Members, types.StringValue(i))
		}

		state.Metadata = &entityMetadataModel{
			UID:         types.StringValue(group.Metadata.UID),
			Etag:        types.StringValue(group.Metadata.Etag),
			Name:        types.StringValue(group.Metadata.Name),
			Namespace:   types.StringValue(group.Metadata.Namespace),
			Title:       types.StringValue(group.Metadata.Title),
			Description: types.StringValue(group.Metadata.Description),
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		}

		for k, v := range group.Metadata.Labels {
			state.Metadata.Labels[k] = v
		}

		for k, v := range group.Metadata.Annotations {
			state.Metadata.Annotations[k] = v
		}

		for _, v := range group.Metadata.Tags {
			state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
		}

		for _, v := range group.Metadata.Links {
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
