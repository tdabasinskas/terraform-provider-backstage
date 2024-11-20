package backstage

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tdabasinskas/go-backstage/v2/backstage"
	"github.com/tdabasinskas/terraform-provider-backstage/internal/transport"
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
	Headers          types.Map    `tfsdk:"headers"`
	Retries          types.Int64  `tfsdk:"retries"`
	TimeoutSeconds   types.Int64  `tfsdk:"timeout_seconds"`
}

const (
	patternURL                 = "https?://.+"
	envBaseURL                 = "BACKSTAGE_BASE_URL"
	envDefaultNamespace        = "BACKSTAGE_DEFAULT_NAMESPACE"
	envHeaders                 = "BACKSTAGE_HEADERS"
	envRetries                 = "BACKSTAGE_RETRIES"
	envTimeoutSeconds          = "BACKSTAGE_TIMEOUT_SECONDS"
	descriptionProviderBaseURL = "Base URL of the Backstage instance, e.g. https://demo.backstage.io. May also be provided via `" + envBaseURL +
		"` environment variable."
	descriptionProviderDefaultNamespace = "Name of default namespace for entities (`default`, if not set). May also be provided via `" + envDefaultNamespace +
		"` environment variable."
	descriptionProviderHeaders = "Headers to be sent with each request to the Backstage API. Useful for authentication. May also be provided via `" + envHeaders +
		"` environment variable."
	descriptionProviderRetries = "Number of retries to attempt on recoverable API errors (default: 0). May also be provided via `" + envRetries +
		"` environment variable."
	descriptionProviderTimeoutSeconds = "Timeout for requests to the Backstage API in seconds (default: 15). May also be provided via `" + envTimeoutSeconds +
		"` environment variable."
)

// Metadata returns the provider type name.
func (p *backstageProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "backstage"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *backstageProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use Backstage provider to interact with the resources supported by [Backstage](https://backstage.io). " +
			"You must configure the provider with proper base URL of your Backstage instance before you can use it.\n\n" +
			"Use the navigation on the left to read about the available resources and data sources.\n\n To learn the basic of Terraform using this provider, " +
			"follow hands-on [get started tutorials](https://learn.hashicorp.com/tutorials/terraform/infrastructure-as-code).\n\n" +
			"Interested in the provider's latest features, or want to make sure you're up to date? Check out the " +
			"[releases](https://github.com/tdabasinskas/terraform-provider-backstage/releases) for version information and release notes.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{Optional: true, MarkdownDescription: descriptionProviderBaseURL, Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(patternURL), "must be a valid URL"),
			}},
			"default_namespace": schema.StringAttribute{Optional: true, MarkdownDescription: descriptionProviderDefaultNamespace, Validators: []validator.String{
				stringvalidator.LengthBetween(1, 63),
				stringvalidator.RegexMatches(regexp.MustCompile(patternEntityName), "must follow Backstage format restrictions"),
			}},
			"headers":         schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: descriptionProviderHeaders},
			"retries":         schema.Int64Attribute{Optional: true, MarkdownDescription: descriptionProviderRetries},
			"timeout_seconds": schema.Int64Attribute{Optional: true, MarkdownDescription: descriptionProviderTimeoutSeconds},
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

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"), "Unknown Base URL of Backstage instance", fmt.Sprintf(
			"Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", envBaseURL))
	}

	if config.DefaultNamespace.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("default_namespace"), "Unknown default entities namespace of Backstage instance", fmt.Sprintf(
			"Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", envDefaultNamespace))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := os.Getenv(envBaseURL)
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if regex := regexp.MustCompile(patternURL); baseURL == "" || !regex.MatchString(baseURL) {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"), "Missing or invalid Base URL of Backstage instance", fmt.Sprintf(
			"The provider cannot create the Backstage API client as there is empty or invalid value for the Backstage Base URL. Set the host value in the "+
				"configuration or use the %s environment variable. If either is already set, ensure the value is not empty and valid.", envBaseURL))

	}

	defaultNamespace := os.Getenv(envDefaultNamespace)
	if !config.DefaultNamespace.IsNull() {
		defaultNamespace = config.DefaultNamespace.ValueString()
	}
	if defaultNamespace == "" {
		defaultNamespace = backstage.DefaultNamespaceName
	}

	if regex := regexp.MustCompile(patternEntityName); !regex.MatchString(defaultNamespace) {
		resp.Diagnostics.AddAttributeError(path.Root("default_namespace"), "Invalid default namespace of Backstage instance", fmt.Sprintf(
			"The provider cannot create the Backstage API client as there is invalid value for the default namespace. Set the host value in the "+
				"configuration or use the %s environment variable. If either is already set, ensure the value is not empty and valid.", envDefaultNamespace))
	}

	headers := make(map[string]string)
	if headersEnv := os.Getenv(envHeaders); headersEnv != "" {
		for _, kv := range regexp.MustCompile(`(.*?)=([^=]*)(?:,|$)`).FindAllStringSubmatch(headersEnv, -1) {
			headers[kv[1]] = kv[2]
		}
	} else {
		if !config.Headers.IsNull() {
			config.Headers.ElementsAs(ctx, &headers, true)
		}
	}

	retries := 0
	retriesStr := os.Getenv(envRetries)
	if retriesStr != "" {
		var err error
		if retries, err = strconv.Atoi(retriesStr); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("retries"), "Invalid number of retries", fmt.Sprintf("The provider cannot create the Backstage API client as there is invalid value for the number of retries: %s.", envRetries))
		}
	} else {
		if !config.Retries.IsNull() {
			retries = int(config.Retries.ValueInt64())
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	timeoutSeconds := 15
	timeoutSecondsStr := os.Getenv(envTimeoutSeconds)
	if timeoutSecondsStr != "" {
		var err error
		if timeoutSeconds, err = strconv.Atoi(timeoutSecondsStr); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("timeout_seconds"), "Invalid timeout for requests to the Backstage API", fmt.Sprintf("The provider cannot create the Backstage API client as there is invalid value for the timeout for requests to the Backstage API: %s.", envTimeoutSeconds))
		}
	} else {
		if !config.TimeoutSeconds.IsNull() {
			timeoutSeconds = int(config.TimeoutSeconds.ValueInt64())
		}
	}

	ctx = tflog.SetField(ctx, "backstage_base_url", baseURL)
	ctx = tflog.SetField(ctx, "backstage_default_namespace", defaultNamespace)
	ctx = tflog.SetField(ctx, "backstage_headers", headers)
	ctx = tflog.SetField(ctx, "backstage_retries", retries)
	ctx = tflog.SetField(ctx, "backstage_timeout_seconds", timeoutSeconds)

	tflog.Debug(ctx, "Creating Backstage API client")

	baseClient := &http.Client{}
	baseClient.Timeout = time.Duration(timeoutSeconds) * time.Second

	if retries > 0 {
		retryableClient := retryablehttp.NewClient()
		retryableClient.RetryMax = retries
		retryableClient.HTTPClient.Timeout = baseClient.Timeout
		baseClient = retryableClient.StandardClient()
	}

	baseClient.Transport = &transport.HeadersTransport{
		BaseTransport: baseClient.Transport,
		Headers:       headers,
	}

	client, err := backstage.NewClient(baseURL, defaultNamespace, baseClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Backstage API client",
			fmt.Sprintf("An unexpected error occurred when creating the Backstage API client: %s", err.Error()),
		)
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *backstageProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLocationResource,
	}
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
