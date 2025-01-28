package provider

import (
	"context"
	"os"

	"github.com/Lela810/pterodactyl-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &pterodactylProvider{}
)

// pterodactylProviderModel maps provider schema data to a Go type.
type pterodactylProviderModel struct {
	Host   types.String `tfsdk:"host"`
	ApiKey types.String `tfsdk:"api_key"`
}

// pterodactylProvider is the provider implementation.
type pterodactylProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pterodactylProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *pterodactylProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pterodactyl"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *pterodactylProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl provider allows Terraform to interact with the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Pterodactyl Panel host URL.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Pterodactyl Panel API key.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *pterodactylProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Pterodactyl Provider")

	// Retrieve provider data from configuration
	var config pterodactylProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Pterodactyl Panel Host",
			"The provider requires a known value for the Pterodactyl Panel host. "+
				"Set the host value in the configuration or use the PTERODACTYL_HOST environment variable.",
		)
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Pterodactyl Panel API Key",
			"The provider requires a known value for the Pterodactyl Panel API key. "+
				"Set the api_key value in the configuration or use the PTERODACTYL_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("PTERODACTYL_HOST")
	apiKey := os.Getenv("PTERODACTYL_API_KEY")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Pterodactyl Panel Host",
			"The provider cannot create the Pterodactyl Panel client as there is a missing or empty value for the Pterodactyl Panel host. "+
				"Set the host value in the configuration or use the PTERODACTYL_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apiKey"),
			"Missing Pterodactyl Panel API Key",
			"The provider cannot create the Pterodactyl Panel client as there is a missing or empty value for the Pterodactyl Panel API key. "+
				"Set the api_key value in the configuration or use the PTERODACTYL_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "pterodactyl_host", host)
	ctx = tflog.SetField(ctx, "pterodactyl_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "pterodactyl_api_key")

	tflog.Debug(ctx, "Creating Pterodactyl client")

	// Create a new Pterodactyl client using the configuration values
	client, err := pterodactyl.NewClient(&host, &apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Pterodactyl API Client",
			"An unexpected error occurred when creating the Pterodactyl API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Pterodactyl Client Error: "+err.Error(),
		)
		return
	}

	// Make the Pterodactyl client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Pterodactyl client created")
}

// DataSources defines the data sources implemented in the provider.
func (p *pterodactylProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// User related data sources
		NewUsersDataSource,
		NewUserDataSource,
		// Node related data sources
		NewNodesDataSource,
		NewNodeDataSource,
		NewNodeAllocationsDataSource,
		// Location related data sources
		NewLocationDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *pterodactylProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewNodeResource,
		NewLocationResource,
	}
}
