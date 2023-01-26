package backstage

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tdabasinskas/go-backstage/backstage"
)

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource is the data source implementation.
type userDataSource struct {
	client *backstage.Client
}

type userDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Namespace  types.String          `tfsdk:"namespace"`
	ApiVersion types.String          `tfsdk:"api_version"`
	Kind       types.String          `tfsdk:"kind"`
	Metadata   *entityMetadataModel  `tfsdk:"metadata"`
	Relations  []entityRelationModel `tfsdk:"relations"`
	Spec       *userSpecModel        `tfsdk:"spec"`
}

type userSpecModel struct {
	Profile  *userSpecProfileModel `tfsdk:"profile"`
	MemberOf []types.String        `tfsdk:"member_of"`
}

type userSpecProfileModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	Email       types.String `tfsdk:"email"`
	Picture     types.String `tfsdk:"picture"`
}

const (
	descriptionUserSpecProfile            = "Profile information about the user, mainly for display purposes."
	descriptionUserSpecProfileDisplayName = "A simple display name to present to users."
	descriptionUserSpecProfileEmail       = "Email where this user can be reached."
	descriptionUserSpecProfilePicture     = "A URL of an image that represents this user."
	descriptionUserSpecMemberOf           = "The list of groups that the user is a direct member of (i.e., no transitive memberships are listed here)."
)

// Metadata returns the data source type name.
func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source.
func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get a specific " +
			"[User entity](https://backstage.io/docs/features/software-catalog/descriptor-format#kind-user) from Backstage Software Catalog.",
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
				"member_of": schema.ListAttribute{Computed: true, Description: descriptionUserSpecMemberOf, ElementType: types.StringType},
				"profile": schema.SingleNestedAttribute{Computed: true, Description: descriptionUserSpecProfile, Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{Computed: true, Description: descriptionUserSpecProfileDisplayName},
					"email":        schema.StringAttribute{Computed: true, Description: descriptionUserSpecProfileEmail},
					"picture":      schema.StringAttribute{Computed: true, Description: descriptionUserSpecProfilePicture},
				}},
			}},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*backstage.Client)
}

// Read refreshes the Terraform state with the latest data.
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Namespace.IsNull() {
		state.Namespace = types.StringValue(backstage.DefaultNamespaceName)
	}

	tflog.Debug(ctx, fmt.Sprintf("Getting User kind %s/%s from Backstage API", state.Name.ValueString(), state.Namespace.ValueString()))
	user, response, err := d.client.Catalog.Users.Get(ctx, state.Name.ValueString(), state.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Backstage User kind",
			fmt.Sprintf("Could not read Backstage User kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), err.Error()),
		)
		return
	}

	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Error reading Backstage User kind",
			fmt.Sprintf("Could not read Backstage User kind %s/%s: %s", state.Namespace.ValueString(), state.Name.ValueString(), response.Status),
		)
		return
	}

	state.ID = types.StringValue(user.Metadata.UID)
	state.ApiVersion = types.StringValue(user.ApiVersion)
	state.Kind = types.StringValue(user.Kind)

	for _, i := range user.Relations {
		state.Relations = append(state.Relations, entityRelationModel{
			Type:      types.StringValue(i.Type),
			TargetRef: types.StringValue(i.TargetRef),
			Target: &entityRelationTargetModel{
				Kind:      types.StringValue(i.Target.Kind),
				Name:      types.StringValue(i.Target.Name),
				Namespace: types.StringValue(i.Target.Namespace)},
		})
	}

	state.Spec = &userSpecModel{
		Profile: &userSpecProfileModel{
			DisplayName: types.StringValue(user.Spec.Profile.DisplayName),
			Email:       types.StringValue(user.Spec.Profile.Email),
			Picture:     types.StringValue(user.Spec.Profile.Picture),
		},
	}

	for _, i := range user.Spec.MemberOf {
		state.Spec.MemberOf = append(state.Spec.MemberOf, types.StringValue(i))
	}

	state.Metadata = &entityMetadataModel{
		UID:         types.StringValue(user.Metadata.UID),
		Etag:        types.StringValue(user.Metadata.Etag),
		Name:        types.StringValue(user.Metadata.Name),
		Namespace:   types.StringValue(user.Metadata.Namespace),
		Title:       types.StringValue(user.Metadata.Title),
		Description: types.StringValue(user.Metadata.Description),
		Annotations: map[string]string{},
		Labels:      map[string]string{},
	}

	for k, v := range user.Metadata.Labels {
		state.Metadata.Labels[k] = v
	}

	for k, v := range user.Metadata.Annotations {
		state.Metadata.Annotations[k] = v
	}

	for _, v := range user.Metadata.Tags {
		state.Metadata.Tags = append(state.Metadata.Tags, types.StringValue(v))
	}

	for _, v := range user.Metadata.Links {
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
