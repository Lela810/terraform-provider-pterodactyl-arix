package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/Lela810/pterodactyl-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

// NewUsersDataSource is a helper function to simplify the provider implementation.
func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// usersDataSource is the data source implementation.
type usersDataSource struct {
	client *pterodactyl.Client
}

// usersDataSourceModel maps the data source schema data.
type usersDataSourceModel struct {
	Users []User `tfsdk:"users"`
}

// User schema data.
type User struct {
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

// Metadata returns the data source type name.
func (d *usersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema defines the schema for the data source.
func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl users data source allows Terraform to read user data from the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description: "The list of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int32Attribute{
							Description: "The ID of the user.",
							Computed:    true,
						},
						"external_id": schema.StringAttribute{
							Description: "The external ID of the user.",
							Computed:    true,
						},
						"uuid": schema.StringAttribute{
							Description: "The UUID of the user.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "The username of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email of the user.",
							Computed:    true,
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
							Description: "The creation date of the user.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "The last update date of the user.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel

	users, err := d.client.GetUsers()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Pterodactyl Users",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, user := range users {
		userState := User{
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

		state.Users = append(state.Users, userState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
