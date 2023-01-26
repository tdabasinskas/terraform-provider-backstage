package backstage

import (
	"context"
	"fmt"
	"os"

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

// Metadata returns the provider type name.
func (p *backstageProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "backstage"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *backstageProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Backstage instance, e.g. https://demo.backstage.io",
				Optional:            true,
			},
			"default_namespace": schema.StringAttribute{
				MarkdownDescription: "Name of default namespace for entities (`default`, if not set)",
				Optional:            true,
			},
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

	const (
		envNameBaseUrl      = "BACKSTAGE_BASE_URL"
		envDefaultNamespace = "BACKSTAGE_DEFAULT_NAMESPACE"
	)

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"),
			"Unknown Base URL of Backstage instance",
			fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, "+
				"or use the %s environment variable.", envNameBaseUrl))
	}

	if config.DefaultNamespace.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("default_namespace"),
			"Unknown default entities namespace of Backstage instance",
			fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, "+
				"or use the %s environment variable.", envDefaultNamespace))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := os.Getenv(envNameBaseUrl)
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	defaultNamespace := os.Getenv(envDefaultNamespace)
	if !config.DefaultNamespace.IsNull() {
		defaultNamespace = config.BaseURL.ValueString()
	}
	if defaultNamespace == "" {
		defaultNamespace = backstage.DefaultNamespaceName
	}

	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"), "Missing Base URL of Backstage instance",
			"The provider cannot create the Backstage API client as there is a missing or empty value for the Backstage base URL. "+
				fmt.Sprintf("Set the base_url value in the configuration or use the %s environment variable. ", envNameBaseUrl)+
				"If either is already set, ensure the value is not empty.",
		)
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

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &backstageProvider{
			version: version,
		}
	}
}
