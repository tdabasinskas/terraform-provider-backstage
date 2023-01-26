package backstage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tdabasinskas/go-backstage/backstage"
)

var _ provider.Provider = &backstageProvider{}

// backstageProvider defines the provider implementation.
type backstageProvider struct {
	version string
}

// backstageProviderModel describes the provider data model.
type backstageProviderModel struct {
	BaseURL          types.String `tfsdk:"base_url"`
	DefaultNamespace types.String `tfsdk:"default_namespace"`
}

const (
	descriptionProviderBaseURL          = "Base URL of the Backstage instance, e.g. https://demo.backstage.io."
	descriptionProviderDefaultNamespace = "Name of default namespace for entities (`default`, if not set)."
)

// Metadata returns the provider type name.
func (p *backstageProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "backstage"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *backstageProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url":          schema.StringAttribute{Required: true, Description: descriptionProviderBaseURL},
			"default_namespace": schema.StringAttribute{Optional: true, Description: descriptionProviderDefaultNamespace},
		},
	}
}

// Configure prepares Backstage API client for data sources and resources.
func (p *backstageProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Backstage API client")

	var config backstageProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := config.BaseURL.ValueString()
	if config.BaseURL.IsUnknown() || config.BaseURL.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"),
			"Unknown Base URL of Backstage instance", "Either target apply the source of the value first, or set the value statically in the configuration")
	}

	if config.DefaultNamespace.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("default_namespace"),
			"Unknown default entities namespace of Backstage instance",
			"Either target apply the source of the value first, or set the value statically in the configuration")
	}

	defaultNamespace := config.DefaultNamespace.ValueString()
	if defaultNamespace == "" {
		defaultNamespace = backstage.DefaultNamespaceName
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "backstage_base_url", baseURL)
	ctx = tflog.SetField(ctx, "backstage_default_namespace", defaultNamespace)

	tflog.Debug(ctx, "Creating Backstage API client")

	client, err := backstage.NewClient(baseURL, defaultNamespace, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Backstage API client",
			fmt.Sprintf("An unexpected error occurred when creating the Backstage API client: %s", err.Error()),
		)
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *backstageProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *backstageProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewEntityDataSource,
		NewApiDataSource,
		NewComponentDataSource,
		NewDomainDataSource,
		NewGroupDataSource,
		NewLocationDataSource,
		NewResourceDataSource,
		NewSystemDataSource,
		NewUserDataSource,
	}
}

// New instantiates a new Backstage provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &backstageProvider{
			version: version,
		}
	}
}
