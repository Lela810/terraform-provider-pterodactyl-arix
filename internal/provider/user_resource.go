package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Lela810/pterodactyl-client-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource is the resource implementation.
type userResource struct {
	client *pterodactyl.Client
}

// userResourceModel maps the resource schema data.
type userResourceModel struct {
	ID        types.Int32  `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl user resource allows Terraform to manage users in the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Description: "The ID of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username of the user.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email of the user.",
				Required:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "The first name of the user.",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "The last name of the user.",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation date of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update date of the user.",
				Computed:    true,
			},
		},
	}
}

// Create a new resource.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create partial user
	partialUser := pterodactyl.PartialUser{
		Username:  plan.Username.ValueString(),
		Email:     plan.Email.ValueString(),
		FirstName: plan.FirstName.ValueString(),
		LastName:  plan.LastName.ValueString(),
	}

	// Create new user
	user, err := r.client.CreateUser(partialUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(user.ID)
	plan.CreatedAt = types.StringValue(user.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed user value from Pterodactyl
	user, err := r.client.GetUser(state.ID.ValueInt32())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pterodactyl User",
			"Could not read Pterodactyl user ID "+strconv.FormatInt(int64(state.ID.ValueInt32()), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Email = types.StringValue(user.Email)
	state.FirstName = types.StringValue(user.FirstName)
	state.LastName = types.StringValue(user.LastName)
	state.UpdatedAt = types.StringValue(user.UpdatedAt.Format(time.RFC3339))
	state.CreatedAt = types.StringValue(user.CreatedAt.Format(time.RFC3339))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create partial user
	var partialUser pterodactyl.PartialUser = pterodactyl.PartialUser{
		Username:  plan.Username.ValueString(),
		Email:     plan.Email.ValueString(),
		FirstName: plan.FirstName.ValueString(),
		LastName:  plan.LastName.ValueString(),
	}

	// Update existing user
	user, err := r.client.UpdateUser(plan.ID.ValueInt32(), partialUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Pterodactyl User",
			"Could not update user, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated values
	plan.Email = types.StringValue(user.Email)
	plan.FirstName = types.StringValue(user.FirstName)
	plan.LastName = types.StringValue(user.LastName)
	plan.UpdatedAt = types.StringValue(user.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing user
	err := r.client.DeleteUser(state.ID.ValueInt32())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Pterodactyl User",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	username := req.ID

	user, err := r.client.GetUserUsername(username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Pterodactyl User",
			"Could not import user: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	state := userResourceModel{
		ID:        types.Int32Value(user.ID),
		Username:  types.StringValue(user.Username),
		Email:     types.StringValue(user.Email),
		FirstName: types.StringValue(user.FirstName),
		LastName:  types.StringValue(user.LastName),
		CreatedAt: types.StringValue(user.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(user.UpdatedAt.Format(time.RFC3339)),
	}

	// Set state to fully populated data
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
