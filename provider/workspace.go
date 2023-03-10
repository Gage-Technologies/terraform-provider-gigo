package provider

import (
	"context"
	"os"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func workspaceDataSource() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information for the active workspace build.",
		ReadContext: func(c context.Context, rd *schema.ResourceData, i interface{}) diag.Diagnostics {
			transition := os.Getenv("GIGO_WORKSPACE_TRANSITION")
			if transition == "" {
				// Default to start!
				transition = "start"
			}
			_ = rd.Set("transition", transition)
			count := 0
			if transition == "start" {
				count = 1
			}
			_ = rd.Set("start_count", count)

			owner := os.Getenv("GIGO_WORKSPACE_OWNER")
			if owner == "" {
				owner = "default"
			}
			_ = rd.Set("owner", owner)

			ownerEmail := os.Getenv("GIGO_WORKSPACE_OWNER_EMAIL")
			_ = rd.Set("owner_email", ownerEmail)

			ownerID := os.Getenv("GIGO_WORKSPACE_OWNER_ID")
			if ownerID == "" {
				ownerID = uuid.Nil.String()
			}
			_ = rd.Set("owner_id", ownerID)

			disk := os.Getenv("GIGO_WORKSPACE_DISK")
			if disk == "" {
				// default to 15GiB
				disk = "15Gi"
			}
			_ = rd.Set("disk", disk)

			cpu := os.Getenv("GIGO_WORKSPACE_CPU")
			if cpu == "" {
				// default to 4 cores
				cpu = "4"
			}
			_ = rd.Set("cpu", cpu)

			mem := os.Getenv("GIGO_WORKSPACE_MEM")
			if mem == "" {
				// default to 4GB
				mem = "4G"
			}
			_ = rd.Set("mem", mem)

			container := os.Getenv("GIGO_WORKSPACE_CONTAINER")
			if container == "" {
				container = "codercom/enterprise-base:ubuntu"
			}
			_ = rd.Set("container", container)

			id := os.Getenv("GIGO_WORKSPACE_ID")
			if id == "" {
				id = uuid.NewString()
			}
			rd.SetId(id)

			config, valid := i.(config)
			if !valid {
				return diag.Errorf("config was unexpected type %q", reflect.TypeOf(i).String())
			}
			rd.Set("access_url", config.URL.String())

			rawPort := config.URL.Port()
			if rawPort == "" {
				rawPort = "80"
				if config.URL.Scheme == "https" {
					rawPort = "443"
				}
			}
			port, err := strconv.Atoi(rawPort)
			if err != nil {
				return diag.Errorf("couldn't parse port %q", port)
			}
			rd.Set("access_port", port)

			return nil
		},
		Schema: map[string]*schema.Schema{
			"access_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The access URL of the Gigo deployment provisioning this workspace.",
			},
			"access_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The access port of the Gigo deployment provisioning this workspace.",
			},
			"start_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: `A computed count based on "transition" state. If "start", count will equal 1.`,
			},
			"transition": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Either "start" or "stop". Use this to start/stop resources with "count".`,
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the workspace owner.",
			},
			"owner_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email address of the workspace owner.",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the workspace owner.",
			},
			"disk": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Disk size of the volume mount.",
			},
			"mem": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Memory size fot the workspace.",
			},
			"cpu": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CPU core count for the workspace.",
			},
			"container": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Container that the workspace will built in.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the workspace.",
			},
		},
	}
}
