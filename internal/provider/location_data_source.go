package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/Lela810/pterodactyl-client-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &locationDataSource{}
	_ datasource.DataSourceWithConfigure = &locationDataSource{}
)

// locationDataSourceModel maps the data source schema data.
type locationDataSourceModel struct {
	ID        types.Int32  `tfsdk:"id"`
	Short     types.String `tfsdk:"short"`
	Long      types.String `tfsdk:"long"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// NewLocationDataSource is a helper function to simplify the provider implementation.
func NewLocationDataSource() datasource.DataSource {
	return &locationDataSource{}
}

// locationDataSource is the data source implementation.
type locationDataSource struct {
	client *pterodactyl.Client
}

// Metadata returns the data source type name.
func (d *locationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

// Schema defines the schema for the data source.
func (d *locationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl location data source allows Terraform to read a location from the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Description: "The ID of the location.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int32{
					int32validator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("short"),
						path.MatchRoot("long"),
					),
				},
			},
			"short": schema.StringAttribute{
				Description: "The short name of the location.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("short"),
						path.MatchRoot("long"),
					),
				},
			},
			"long": schema.StringAttribute{
				Description: "The long name of the location.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("short"),
						path.MatchRoot("long"),
					),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The date and time the location was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The date and time the location was last updated.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *locationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state locationDataSourceModel

	// Get the attributes from the request
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var location pterodactyl.Location

	if !state.ID.IsNull() {
		var err error
		location, err = d.client.GetLocation(state.ID.ValueInt32())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Pterodactyl Location",
				err.Error(),
			)
			return
		}
	} else if !state.Short.IsNull() {
		locations, err := d.client.GetLocations()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Pterodactyl Locations",
				err.Error(),
			)
		}

		for _, loc := range locations {
			if loc.Short != state.Short.ValueString() {
				continue
			}
			location = loc
			break
		}
	} else if !state.Long.IsNull() {
		locations, err := d.client.GetLocations()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Pterodactyl Locations",
				err.Error(),
			)
		}

		for _, loc := range locations {
			if loc.Long != state.Long.ValueString() {
				continue
			}
			location = loc
			break
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Attribute",
			"One of 'id', 'short' or 'long' must be specified.",
		)
		return
	}

	// Map response body to model
	state = locationDataSourceModel{
		ID:        types.Int32Value(location.ID),
		Short:     types.StringValue(location.Short),
		Long:      types.StringValue(location.Long),
		CreatedAt: types.StringValue(location.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(location.UpdatedAt.Format(time.RFC3339)),
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *locationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pterodactyl.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *pterodactyl.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
