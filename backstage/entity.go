package backstage

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type entityModel struct {
	ApiVersion types.String           `tfsdk:"api_version"`
	Kind       types.String           `tfsdk:"kind"`
	Metadata   *entityMetadataModel   `tfsdk:"metadata"`
	Spec       map[string]interface{} `tfsdk:"spec"`
	Relations  []entityRelationModel  `tfsdk:"relations"`
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
