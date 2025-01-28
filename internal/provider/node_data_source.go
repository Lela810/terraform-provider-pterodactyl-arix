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
	_ datasource.DataSource              = &nodeDataSource{}
	_ datasource.DataSourceWithConfigure = &nodeDataSource{}
)

// nodeDataSourceModel maps the data source schema data.
type nodeDataSourceModel struct {
	ID                 types.Int32  `tfsdk:"id"`
	UUID               types.String `tfsdk:"uuid"`
	Public             types.Bool   `tfsdk:"public"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ContainerText      types.String `tfsdk:"container_text"`
	LocationID         types.Int32  `tfsdk:"location_id"`
	FQDN               types.String `tfsdk:"fqdn"`
	Scheme             types.String `tfsdk:"scheme"`
	BehindProxy        types.Bool   `tfsdk:"behind_proxy"`
	MaintenanceMode    types.Bool   `tfsdk:"maintenance_mode"`
	Memory             types.Int32  `tfsdk:"memory"`
	MemoryOverallocate types.Int32  `tfsdk:"memory_overallocate"`
	Disk               types.Int32  `tfsdk:"disk"`
	DiskOverallocate   types.Int32  `tfsdk:"disk_overallocate"`
	UploadSize         types.Int32  `tfsdk:"upload_size"`
	DaemonListen       types.Int32  `tfsdk:"daemon_listen"`
	DaemonText         types.String `tfsdk:"daemon_text"`
	DaemonSFTP         types.Int32  `tfsdk:"daemon_sftp"`
	DaemonBase         types.String `tfsdk:"daemon_base"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

// NewNodeDataSource is a helper function to simplify the provider implementation.
func NewNodeDataSource() datasource.DataSource {
	return &nodeDataSource{}
}

// nodeDataSource is the data source implementation.
type nodeDataSource struct {
	client *pterodactyl.Client
}

// Metadata returns the data source type name.
func (d *nodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

// Schema defines the schema for the data source.
func (d *nodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Pterodactyl node data source allows Terraform to read a nodes data from the Pterodactyl Panel API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Description: "The ID of the node.",
				Computed:    true,
				Optional:    true,
				Validators: []validator.Int32{
					int32validator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("uuid"),
						path.MatchRoot("name"),
					),
				},
			},
			"uuid": schema.StringAttribute{
				Description: "The UUID of the node.",
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("uuid"),
						path.MatchRoot("name"),
					),
				},
			},
			"public": schema.BoolAttribute{
				Description: "The public status of the node.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the node.",
				Computed:    true,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("uuid"),
						path.MatchRoot("name"),
					),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the node.",
				Computed:    true,
			},
			"container_text": schema.StringAttribute{
				Description: "The name of the container inside the console.",
				Computed:    true,
			},
			"location_id": schema.Int32Attribute{
				Description: "The location ID of the node.",
				Computed:    true,
			},
			"fqdn": schema.StringAttribute{
				Description: "The FQDN of the node.",
				Computed:    true,
			},
			"scheme": schema.StringAttribute{
				Description: "The scheme of the node.",
				Computed:    true,
			},
			"behind_proxy": schema.BoolAttribute{
				Description: "The behind proxy status of the node.",
				Computed:    true,
			},
			"maintenance_mode": schema.BoolAttribute{
				Description: "The maintenance mode status of the node.",
				Computed:    true,
			},
			"memory": schema.Int32Attribute{
				Description: "The memory of the node.",
				Computed:    true,
			},
			"memory_overallocate": schema.Int32Attribute{
				Description: "The memory overallocate of the node.",
				Computed:    true,
			},
			"disk": schema.Int32Attribute{
				Description: "The disk of the node.",
				Computed:    true,
			},
			"disk_overallocate": schema.Int32Attribute{
				Description: "The disk overallocate of the node.",
				Computed:    true,
			},
			"upload_size": schema.Int32Attribute{
				Description: "The upload size of the node.",
				Computed:    true,
			},
			"daemon_listen": schema.Int32Attribute{
				Description: "The daemon listen of the node.",
				Computed:    true,
			},
			"daemon_text": schema.StringAttribute{
				Description: "The name of the daemon inside the console.",
				Computed:    true,
			},
			"daemon_sftp": schema.Int32Attribute{
				Description: "The daemon SFTP of the node.",
				Computed:    true,
			},
			"daemon_base": schema.StringAttribute{
				Description: "The file base of the daemon of the node",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation date of the node.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update date of the node.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *nodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeDataSourceModel

	// Get the attributes from the request
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the node from the API based on the provided attribute
	var node pterodactyl.Node
	if !state.ID.IsNull() {
		var err error
		node, err = d.client.GetNode(state.ID.ValueInt32())

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Pterodactyl Node",
				err.Error(),
			)
			return
		}
	} else if !state.UUID.IsNull() {
		uuid := state.UUID.ValueString()
		nodes, err := d.client.GetNodes()

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Pterodactyl Node",
				err.Error(),
			)
			return
		}

		for _, n := range nodes {
			if n.UUID == uuid {
				node = n
				break
			}
		}
	} else if !state.Name.IsNull() {
		name := state.Name.ValueString()
		nodes, err := d.client.GetNodes()

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Pterodactyl Node",
				err.Error(),
			)
			return
		}

		for _, n := range nodes {
			if n.Name == name {
				node = n
				break
			}
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Attribute",
			"One of 'id', 'uuid' or 'name' must be specified.",
		)
		return
	}

	// Map response body to model
	state = nodeDataSourceModel{
		ID:                 types.Int32Value(node.ID),
		UUID:               types.StringValue(node.UUID),
		Public:             types.BoolValue(node.Public),
		Name:               types.StringValue(node.Name),
		Description:        types.StringValue(node.Description),
		ContainerText:      types.StringValue(node.ContainerText),
		LocationID:         types.Int32Value(node.LocationID),
		FQDN:               types.StringValue(node.FQDN),
		Scheme:             types.StringValue(node.Scheme),
		BehindProxy:        types.BoolValue(node.BehindProxy),
		MaintenanceMode:    types.BoolValue(node.MaintenanceMode),
		Memory:             types.Int32Value(node.Memory),
		MemoryOverallocate: types.Int32Value(node.MemoryOverallocate),
		Disk:               types.Int32Value(node.Disk),
		DiskOverallocate:   types.Int32Value(node.DiskOverallocate),
		UploadSize:         types.Int32Value(node.UploadSize),
		DaemonListen:       types.Int32Value(node.DaemonListen),
		DaemonText:         types.StringValue(node.DaemonText),
		DaemonSFTP:         types.Int32Value(node.DaemonSFTP),
		DaemonBase:         types.StringValue(node.DaemonBase),
		CreatedAt:          types.StringValue(node.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:          types.StringValue(node.UpdatedAt.Format(time.RFC3339)),
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *nodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
