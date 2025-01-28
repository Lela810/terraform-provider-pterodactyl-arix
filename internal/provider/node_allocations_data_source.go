package provider

import (
	"context"
	"fmt"

	"github.com/Lela810/pterodactyl-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &nodeAllocationsDataSource{}
	_ datasource.DataSourceWithConfigure = &nodeAllocationsDataSource{}
)

// NewNodeAllocationsDataSource is a helper function to simplify the provider implementation.
func NewNodeAllocationsDataSource() datasource.DataSource {
	return &nodeAllocationsDataSource{}
}

// nodeAllocationsDataSource is the data source implementation.
type nodeAllocationsDataSource struct {
	client *pterodactyl.Client
}

// nodeAllocationsDataSourceModel maps the data source schema data.
type nodeAllocationsDataSourceModel struct {
	NodeID          int32        `tfsdk:"nodeid"`
	NodeAllocations []Allocation `tfsdk:"allocations"`
}

// Allocation schema data.
type Allocation struct {
	ID       types.Int32  `tfsdk:"id"`
	IP       types.String `tfsdk:"ip"`
	Alias    types.String `tfsdk:"alias"`
	Port     types.Int32  `tfsdk:"port"`
	Notes    types.String `tfsdk:"notes"`
	Assigned types.Bool   `tfsdk:"assigned"`
}

// Metadata returns the data source type name.
func (d *nodeAllocationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_allocations"
}

// Schema defines the schema for the data source.
func (d *nodeAllocationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl IP Allocations for servers.",
		Attributes: map[string]schema.Attribute{
			"nodeid": schema.Int32Attribute{
				Description: "The ID of the node to get allocations from.",
				Required:    true,
			},
			"allocations": schema.ListNestedAttribute{
				Description: "The list of allocations to a node.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int32Attribute{
							Description: "The ID of the node.",
							Computed:    true,
						},
						"ip": schema.StringAttribute{
							Description: "The IP that is allocated",
							Computed:    true,
						},
						"alias": schema.StringAttribute{
							Description: "A alias for the allocation",
							Computed:    true,
						},
						"port": schema.Int32Attribute{
							Description: "The port allocated in the allocation",
							Computed:    true,
						},
						"notes": schema.StringAttribute{
							Description: "Any notes to the allocation",
							Computed:    true,
						},
						"assigned": schema.BoolAttribute{
							Description: "Is the allocation assigned?",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *nodeAllocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeAllocationsDataSourceModel

	// Get the attributes from the request
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes, err := d.client.GetNodeAllocations(state.NodeID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Pterodactyl Nodes",
			err.Error(),
		)
		return
	}

	state.NodeAllocations = make([]Allocation, len(nodes))

	// Map response body to model
	for i, allocation := range nodes {
		state.NodeAllocations[i] = Allocation{
			ID:       types.Int32Value(allocation.ID),
			IP:       types.StringValue(allocation.IP),
			Alias:    types.StringValue(allocation.Alias),
			Port:     types.Int32Value(allocation.Port),
			Notes:    types.StringValue(allocation.Notes),
			Assigned: types.BoolValue(allocation.Assigned),
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *nodeAllocationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
