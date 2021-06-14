package cloud

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/minectl/pgk/automation"
	"github.com/pulumi/pulumi-cloudinit/sdk/go/cloudinit"
	do "github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type DigitalOcean struct {
}

func NewDigitalOcean() *DigitalOcean {
	do := &DigitalOcean{}
	do.InstallPlugin()
	return do
}

func (d *DigitalOcean) InstallPlugin() error {
	ctx := context.Background()
	w, err := auto.NewLocalWorkspace(ctx)
	if err != nil {
		log.Println("Failed to setup and run http server: %v\n", err)
		os.Exit(1)
	}
	err = w.InstallPlugin(ctx, "digitalocean", "4.4.1")
	if err != nil {
		log.Println("Failed to install plugin: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func (d *DigitalOcean) UpdateServer(args automation.ServerArgs) (*automation.RessourceResults, error) {
	log.Print("UpdateServer")
	return executeDOPulumiProgram(args, d, true)
}

func (d *DigitalOcean) CreateServer(args automation.ServerArgs) (*automation.RessourceResults, error) {
	log.Print("CreateServer")
	return executeDOPulumiProgram(args, d, false)
}

func executeDOPulumiProgram(args automation.ServerArgs, d *DigitalOcean, update bool) (*automation.RessourceResults, error) {
	ctx := context.Background()
	program := d.Bootstrap(args)
	var s auto.Stack
	var err error
	if update {
		log.Print("SelectStackInlineSource")
		s, err = auto.SelectStackInlineSource(ctx, args.StackName, "Project", program)
	} else {
		log.Print("NewStackInlineSource")
		s, err = auto.NewStackInlineSource(ctx, args.StackName, "Project", program)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Print("Up")
	upRes, err := s.Up(ctx, optup.ProgressStreams(os.Stdout))
	if err != nil {
		log.Println(err)
		s.Destroy(ctx, optdestroy.ProgressStreams(os.Stdout))
		err = s.Workspace().RemoveStack(ctx, args.StackName)
		return nil, err
	}
	return &automation.RessourceResults{
		PublicIP: upRes.Outputs["minecraft-public"].Value.(string),
	}, nil
}

func (d *DigitalOcean) DeleteServer(id string) error {
	log.Println("DeleteServer")
	ctx := context.Background()
	program := d.Bootstrap(automation.ServerArgs{
		StackName: id,
		Size:      "",
		Region:    "",
	})
	s, err := auto.SelectStackInlineSource(ctx, id, "Project", program)
	if err != nil {
		return err
	}
	_, err = s.Destroy(ctx, optdestroy.ProgressStreams(os.Stdout))
	if err != nil {
		return err
	}
	err = s.Workspace().RemoveStack(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *DigitalOcean) Bootstrap(args automation.ServerArgs) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		fmt.Println(args)

		conf, err := cloudinit.NewConfig(ctx, "test", &cloudinit.ConfigArgs{
			Gzip:         pulumi.Bool(false),
			Base64Encode: pulumi.Bool(false),
			Parts: cloudinit.ConfigPartArray{
				&cloudinit.ConfigPartArgs{
					Content:     pulumi.String(strings.Replace(cloudConfig, "vdc", "sda", -1)),
					ContentType: pulumi.String("text/cloud-config"),
				},
			},
		})
		if err != nil {
			return err
		}
		pubKeyFile, err := ioutil.ReadFile(args.SSH)
		if err != nil {
			return err
		}

		volume, err := do.NewVolume(ctx, "minecraft-vol", &do.VolumeArgs{
			Region:                pulumi.String(args.Region),
			Size:                  pulumi.Int(args.VolumeSize),
			InitialFilesystemType: pulumi.String("ext4"),
			Description:           pulumi.String("volume for storing the minecraft data"),
		})
		if err != nil {
			return err
		}

		sshPubKey, err := do.NewSshKey(ctx, "minecraft-ssh", &do.SshKeyArgs{
			PublicKey: pulumi.String(pubKeyFile),
		})
		droplet, err := do.NewDroplet(ctx, args.StackName, &do.DropletArgs{
			Name:   pulumi.String(args.StackName),
			Image:  pulumi.String("ubuntu-20-10-x64"),
			Region: pulumi.String(args.Region),
			Size:   pulumi.String(args.Size),
			VolumeIds: pulumi.StringArray{
				volume.ID(),
			},
			UserData: conf.Rendered,
			SshKeys: pulumi.StringArray{
				sshPubKey.Fingerprint,
			},
		})

		if err != nil {
			return err
		}

		fip, err := do.NewFloatingIp(ctx, "floating", &do.FloatingIpArgs{
			Region: droplet.Region,
		})

		_ = droplet.Name.ApplyT(func(id string) error {
			lookup, err := do.LookupDroplet(ctx, &do.LookupDropletArgs{
				Name: &id,
			}, nil)
			if err != nil {
				return err
			}
			_, err = do.NewFloatingIpAssignment(ctx, "floating-ip-assignment", &do.FloatingIpAssignmentArgs{
				IpAddress: fip.IpAddress,
				DropletId: pulumi.Int(lookup.Id),
			})
			if err != nil {
				return err
			}
			return nil
		}).(pulumi.AnyOutput)

		ctx.Export("fingerprint", sshPubKey.Fingerprint)
		ctx.Export("minecraft-public", fip.IpAddress)
		ctx.Export("minecraft-id", droplet.ID())
		return nil
	}
}
