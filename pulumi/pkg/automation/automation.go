package automation

import (
	"context"
	"github.com/pulumi/pulumi-openstack/sdk/v3/go/openstack/blockstorage"
	"github.com/pulumi/pulumi-openstack/sdk/v3/go/openstack/compute"
	"github.com/pulumi/pulumi-openstack/sdk/v3/go/openstack/networking"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"log"
	"os"
)

const (
	Project = "minescale"
)

type Automation interface {
	ListServer()
	CreateServer(name, flavor, cidr, pubkey string) (string, error)
}

type AutomationImpl struct {
}

func (a AutomationImpl) CreateServer(name, flavor, cidr, pubkey string) (string, error) {
	log.Print("CreateServer")
	ctx := context.Background()

	program := createPulumiProgram(name, flavor, cidr, pubkey)
	log.Print("createPulumiProgram")
	s, err := auto.NewStackInlineSource(ctx, name, Project, program)
	log.Print("NewStackInlineSource")
	if err != nil {
		log.Println(err)
		return "", err
	}
	upRes, err := s.Up(ctx, optup.ProgressStreams(os.Stdout))
	log.Print("Up")
	return upRes.Outputs["minecraft-public"].Value.(string), nil
}

func NewAutomationProgram() *AutomationImpl {
	p := &AutomationImpl{}
	openstackPlugin()
	return p
}

func (a AutomationImpl) ListServer() {
	ctx := context.Background()

	ws, err := auto.NewLocalWorkspace(ctx, auto.Project(workspace.Project{
		Name:    tokens.PackageName(Project),
		Runtime: workspace.NewProjectRuntimeInfo("go", nil),
	}))
	if err != nil {
		log.Println("%v\n", err)
		return
	}
	stacks, err := ws.ListStacks(ctx)
	if err != nil {
		log.Println("%v\n", err)
		return
	}
	var ids []string
	for _, stack := range stacks {
		ids = append(ids, stack.Name)
	}
	log.Printf("%s", ids)
}

func createPulumiProgram(name, flavor, cidr, pubkey string) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		keypair, err := compute.NewKeypair(ctx, "minecraft-keypair", &compute.KeypairArgs{
			Name:      pulumi.Sprintf("%s-kp", name),
			PublicKey: pulumi.String(pubkey),
		})
		network, err := networking.NewNetwork(ctx, "minecraft-net", &networking.NetworkArgs{
			Name:         pulumi.Sprintf("%s-net", name),
			AdminStateUp: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		subnet, err := networking.NewSubnet(ctx, "minecraft-snet", &networking.SubnetArgs{
			Name:      pulumi.Sprintf("%s-snet", name),
			NetworkId: network.ID(),
			Cidr:      pulumi.String(cidr),
			IpVersion: pulumi.Int(4),
			DnsNameservers: pulumi.StringArray{
				pulumi.String("8.8.8.8"),
				pulumi.String("8.8.4.4"),
			},
		})
		if err != nil {
			return err
		}

		pool := "floating-net"
		floating, err := networking.LookupNetwork(ctx, &networking.LookupNetworkArgs{
			Name: &pool,
		}, nil)

		router, err := networking.NewRouter(ctx, "minecraft-router", &networking.RouterArgs{
			Name:              pulumi.Sprintf("%s-router", name),
			AdminStateUp:      pulumi.Bool(true),
			ExternalNetworkId: pulumi.String(floating.Id),
		})
		if err != nil {
			return err
		}
		_, err = networking.NewRouterInterface(ctx, "minecraft-ri", &networking.RouterInterfaceArgs{
			RouterId: router.ID(),
			SubnetId: subnet.ID(),
		})
		if err != nil {
			return err
		}
		secgroup, err := networking.NewSecGroup(ctx, "minecraft-sg", &networking.SecGroupArgs{
			Name:        pulumi.Sprintf("%s-sg", name),
			Description: pulumi.String("Security group for the Terraform nodes instances"),
		})
		if err != nil {
			return err
		}
		_, err = networking.NewSecGroupRule(ctx, "minecraft-22-sgr", &networking.SecGroupRuleArgs{
			Direction:       pulumi.String("ingress"),
			Ethertype:       pulumi.String("IPv4"),
			Protocol:        pulumi.String("tcp"),
			PortRangeMin:    pulumi.Int(22),
			PortRangeMax:    pulumi.Int(22),
			RemoteIpPrefix:  pulumi.String("0.0.0.0/0"),
			SecurityGroupId: secgroup.ID(),
		})
		if err != nil {
			return err
		}
		_, err = networking.NewSecGroupRule(ctx, "minecraft-19132-sgr", &networking.SecGroupRuleArgs{
			Direction:       pulumi.String("ingress"),
			Ethertype:       pulumi.String("IPv4"),
			Protocol:        pulumi.String("tcp"),
			PortRangeMin:    pulumi.Int(19132),
			PortRangeMax:    pulumi.Int(19132),
			RemoteIpPrefix:  pulumi.String("0.0.0.0/0"),
			SecurityGroupId: secgroup.ID(),
		})
		if err != nil {
			return err
		}
		vm, err := compute.NewInstance(ctx, "minecraft-vm", &compute.InstanceArgs{
			Name:       pulumi.Sprintf("%s-router", name),
			FlavorName: pulumi.String(flavor),
			KeyPair:    keypair.Name,
			SecurityGroups: pulumi.StringArray{
				pulumi.String("default"),
				secgroup.Name,
			},
			//UserData: pulumi.Any(_var.User_data),
			Networks: compute.InstanceNetworkArray{
				&compute.InstanceNetworkArgs{
					Name: network.Name,
				},
			},
			BlockDevices: compute.InstanceBlockDeviceArray{
				&compute.InstanceBlockDeviceArgs{
					Uuid:                pulumi.String("b017f5da-86e2-49ec-98ce-14250f758bfa"),
					SourceType:          pulumi.String("image"),
					BootIndex:           pulumi.Int(0),
					DestinationType:     pulumi.String("volume"),
					VolumeSize:          pulumi.Int(10),
					DeleteOnTermination: pulumi.Bool(true),
				},
			},
		})
		if err != nil {
			return err
		}
		volume, err := blockstorage.NewVolume(ctx, "minecraft-vl", &blockstorage.VolumeArgs{
			Name:        pulumi.Sprintf("%s-vl", name),
			Description: pulumi.String("Storage for the minecraft server"),
			Size:        pulumi.Int(500),
			VolumeType:  pulumi.String("storage_premium_perf2"),
		})
		if err != nil {
			return err
		}
		_, err = compute.NewVolumeAttach(ctx, "t4ch_volume_attach", &compute.VolumeAttachArgs{
			VolumeId:   volume.ID(),
			InstanceId: vm.ID(),
		})
		if err != nil {
			return err
		}
		fip, err := networking.NewFloatingIp(ctx, "minecraft_fip", &networking.FloatingIpArgs{
			Pool: pulumi.String(pool),
		})
		if err != nil {
			return err
		}
		_, err = compute.NewFloatingIpAssociate(ctx, "minecraft_fipa", &compute.FloatingIpAssociateArgs{
			InstanceId: vm.ID(),
			FloatingIp: fip.Address,
		})
		if err != nil {
			return err
		}

		ctx.Export("minecraft-public", fip.Address)
		return nil
	}
}

func openstackPlugin() {
	ctx := context.Background()
	w, err := auto.NewLocalWorkspace(ctx)
	if err != nil {
		log.Println("Failed to setup and run http server: %v\n", err)
		os.Exit(1)
	}
	err = w.InstallPlugin(ctx, "openstack", "3.2.0")
	if err != nil {
		log.Println("Failed to install plugin: %v\n", err)
		os.Exit(1)
	}
}
