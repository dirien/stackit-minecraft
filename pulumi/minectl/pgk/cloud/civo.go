package cloud

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/minectl/pgk/automation"
	"github.com/pulumi/pulumi-civo/sdk/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io/ioutil"
	"log"
	"os"
)

//go:embed civo.sh
var bash string

type Civo struct {
}

func NewCivo() *Civo {
	do := &Civo{}
	do.InstallPlugin()
	return do
}

func (c *Civo) CreateServer(args automation.ServerArgs) (*automation.RessourceResults, error) {
	log.Print("CreateServer")
	return executeCivoPulumiProgram(args, c, false)
}

func (c *Civo) DeleteServer(id string) error {
	log.Println("DeleteServer")
	ctx := context.Background()
	program := c.Bootstrap(automation.ServerArgs{
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

func (c *Civo) UpdateServer(args automation.ServerArgs) (*automation.RessourceResults, error) {
	log.Print("UpdateServer")
	return executeCivoPulumiProgram(args, c, true)
}

func executeCivoPulumiProgram(args automation.ServerArgs, d *Civo, update bool) (*automation.RessourceResults, error) {
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

func (c *Civo) InstallPlugin() error {
	ctx := context.Background()
	w, err := auto.NewLocalWorkspace(ctx)
	if err != nil {
		log.Println("Failed to setup and run http server: %v\n", err)
		os.Exit(1)
	}
	err = w.InstallPlugin(ctx, "civo", "1.0.1")
	if err != nil {
		log.Println("Failed to install plugin: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func (c *Civo) Bootstrap(args automation.ServerArgs) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		fmt.Println(args)

		pubKeyFile, err := ioutil.ReadFile(args.SSH)
		if err != nil {
			return err
		}
		log.Print("sshPubKey")
		sshPubKey, err := civo.NewSshKey(ctx, "ssh-keys", &civo.SshKeyArgs{
			Name:      pulumi.String("minecraft-sshkey"),
			PublicKey: pulumi.String(pubKeyFile),
		})
		if err != nil {
			return err
		}
		var sort = "asc"
		result, err := civo.LookupTemplate(ctx, &civo.LookupTemplateArgs{
			Filters: []civo.GetTemplateFilter{
				{
					Key:    "name",
					Values: []string{"ubuntu-focal"},
				},
			},
			Sorts: []civo.GetTemplateSort{
				{
					Key:       "version",
					Direction: &sort,
				},
			},
		})
		if err != nil {
			return err
		}
		log.Print("template")

		log.Print("instance")
		instance, err := civo.NewInstance(ctx, args.StackName, &civo.InstanceArgs{
			Hostname:         pulumi.String(args.StackName),
			Size:             pulumi.String(args.Size),
			Region:           pulumi.String(args.Region),
			SshkeyId:         sshPubKey.ID(),
			PublicIpRequired: pulumi.String("create"),
			InitialUser:      pulumi.String("root"),
			Template:         pulumi.String(result.Templates[0].Id),
			Script:           pulumi.String(bash),
		})
		if err != nil {
			return err
		}

		ctx.Export("fingerprint", sshPubKey.Fingerprint)
		ctx.Export("template", pulumi.String(result.Templates[0].Id))
		ctx.Export("minecraft-public", instance.PublicIp)
		ctx.Export("minecraft-id", instance.ID())
		return nil
	}
}
