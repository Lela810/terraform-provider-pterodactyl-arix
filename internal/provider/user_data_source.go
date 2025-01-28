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
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// userDataSourceModel maps the data source schema data.
type userDataSourceModel struct {
	ID         types.Int32  `tfsdk:"id"`
	ExternalID types.String `tfsdk:"external_id"`
	UUID       types.String `tfsdk:"uuid"`
	Username   types.String `tfsdk:"username"`
	Email      types.String `tfsdk:"email"`
	FirstName  types.String `tfsdk:"first_name"`
	LastName   types.String `tfsdk:"last_name"`
	Language   types.String `tfsdk:"language"`
	RootAdmin  types.Bool   `tfsdk:"root_admin"`
	Is2FA      types.Bool   `tfsdk:"is_2fa"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource is the data source implementation.
type userDataSource struct {
	client *pterodactyl.Client
}

// Metadata returns the data source type name.
func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source.
func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl user data source allows Terraform to read user data from the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Description: "The ID of the user.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int32{
					int32validator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("external_id"),
						path.MatchRoot("username"),
						path.MatchRoot("email"),
					),
				},
			},
			"external_id": schema.StringAttribute{
				Description: "The external ID of the user.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("external_id"),
						path.MatchRoot("username"),
						path.MatchRoot("email"),
					),
				},
			},
			"uuid": schema.StringAttribute{
				Description: "The UUID of the user.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username of the user.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("external_id"),
						path.MatchRoot("username"),
						path.MatchRoot("email"),
					),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email of the user.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("external_id"),
						path.MatchRoot("username"),
						path.MatchRoot("email"),
					),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the user.",
				Computed:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user.",
				Computed:    true,
			},
			"language": schema.StringAttribute{
				Description: "The language of the user.",
				Computed:    true,
			},
			"root_admin": schema.BoolAttribute{
				Description: "Is the user the root admin.",
				Computed:    true,
			},
			"is_2fa": schema.BoolAttribute{
				Description: "Is the user using 2FA.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The date and time the user was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The date and time the user was last updated.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDataSourceModel

	// Get the attributes from the request
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the user from the API based on the provided attribute
	var user pterodactyl.User
	var err error
	if !state.ID.IsNull() {
		user, err = d.client.GetUser(state.ID.ValueInt32())
	} else if !state.Username.IsNull() {
		user, err = d.client.GetUserUsername(state.Username.ValueString())
	} else if !state.Email.IsNull() {
		user, err = d.client.GetUserEmail(state.Email.ValueString())
	} else if !state.ExternalID.IsNull() {
		user, err = d.client.GetUserExternalID(state.ExternalID.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Attribute",
			"One of 'id', 'username', or 'email' must be specified.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Pterodactyl User",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state = userDataSourceModel{
		ID:         types.Int32Value(user.ID),
		ExternalID: types.StringValue(user.ExternalID),
		UUID:       types.StringValue(user.UUID),
		Username:   types.StringValue(user.Username),
		Email:      types.StringValue(user.Email),
		FirstName:  types.StringValue(user.FirstName),
		LastName:   types.StringValue(user.LastName),
		Language:   types.StringValue(user.Language),
		RootAdmin:  types.BoolValue(user.RootAdmin),
		Is2FA:      types.BoolValue(user.Is2FA),
		CreatedAt:  types.StringValue(user.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:  types.StringValue(user.UpdatedAt.Format(time.RFC3339)),
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
