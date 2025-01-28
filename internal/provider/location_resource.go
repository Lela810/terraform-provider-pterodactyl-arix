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
	_ resource.Resource                = &locationResource{}
	_ resource.ResourceWithConfigure   = &locationResource{}
	_ resource.ResourceWithImportState = &locationResource{}
)

// NewLocationResource is a helper function to simplify the provider implementation.
func NewLocationResource() resource.Resource {
	return &locationResource{}
}

// locationResource is the resource implementation.
type locationResource struct {
	client *pterodactyl.Client
}

// locationResourceModel maps the resource schema data.
type locationResourceModel struct {
	ID        types.Int32  `tfsdk:"id"`
	Short     types.String `tfsdk:"short"`
	Long      types.String `tfsdk:"long"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *locationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

// Schema defines the schema for the resource.
func (r *locationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl location resource allows Terraform to manage locations in the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Description: "The ID of the location.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"short": schema.StringAttribute{
				Description: "The short name of the location.",
				Required:    true,
			},
			"long": schema.StringAttribute{
				Description: "The long name of the location.",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation date of the location.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update date of the location.",
				Computed:    true,
			},
		},
	}
}

// Create a new resource.
func (r *locationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan locationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create partial location
	partialLocation := pterodactyl.PartialLocation{
		Short: plan.Short.ValueString(),
		Long:  plan.Long.ValueString(),
	}

	// Create new location
	location, err := r.client.CreateLocation(partialLocation)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating location",
			"Could not create location, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.Int32Value(location.ID)
	plan.CreatedAt = types.StringValue(location.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *locationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state locationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed location value from Pterodactyl
	location, err := r.client.GetLocation(state.ID.ValueInt32())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pterodactyl Location",
			"Could not read Pterodactyl location ID "+strconv.FormatInt(int64(state.ID.ValueInt32()), 10)+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Short = types.StringValue(location.Short)
	state.Long = types.StringValue(location.Long)
	state.UpdatedAt = types.StringValue(location.UpdatedAt.Format(time.RFC3339))
	state.CreatedAt = types.StringValue(location.CreatedAt.Format(time.RFC3339))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *locationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan locationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create partial location
	var partialLocation = pterodactyl.PartialLocation{
		Short: plan.Short.ValueString(),
		Long:  plan.Long.ValueString(),
	}

	// Update existing location
	location, err := r.client.UpdateLocation(plan.ID.ValueInt32(), partialLocation)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Pterodactyl Location",
			"Could not update location, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated values
	plan.Short = types.StringValue(location.Short)
	plan.Long = types.StringValue(location.Long)
	plan.UpdatedAt = types.StringValue(location.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *locationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state locationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing location
	err := r.client.DeleteLocation(state.ID.ValueInt32())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Pterodactyl Location",
			"Could not delete location, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *locationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *locationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	locationID, err := strconv.ParseInt(req.ID, 10, 32)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing state",
			"Couldn't convert id to int",
		)
	}

	location, err := r.client.GetLocation(int32(locationID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Pterodactyl Location",
			"Could not import location: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	state := locationResourceModel{
		ID:        types.Int32Value(location.ID),
		Long:      types.StringValue(location.Long),
		Short:     types.StringValue(location.Short),
		CreatedAt: types.StringValue(location.CreatedAt.Format(time.RFC3339)),
		UpdatedAt: types.StringValue(location.UpdatedAt.Format(time.RFC3339)),
	}

	// Set state to fully populated data
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
