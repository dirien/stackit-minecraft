package provisioner

import (
	"github.com/minectl/pgk/automation"
	"github.com/minectl/pgk/cloud"
	"github.com/minectl/pgk/common"
	"github.com/minectl/pgk/manifest"
)

type Provisioner interface {
	CreateServer() (*automation.RessourceResults, error)
	DeleteServer() error
	UpdateServer() (*automation.RessourceResults, error)
}

type PulumiProvisioner struct {
	auto     automation.Automation
	manifest *manifest.MinecraftServer
	args     automation.ServerArgs
}

func (p PulumiProvisioner) UpdateServer() (*automation.RessourceResults, error) {
	return p.auto.UpdateServer(p.args)
}

func (p PulumiProvisioner) CreateServer() (*automation.RessourceResults, error) {
	return p.auto.CreateServer(p.args)
}

func (p PulumiProvisioner) DeleteServer() error {
	return p.auto.DeleteServer(p.args.StackName)
}

func NewProvisioner(manifestPath string) *PulumiProvisioner {
	manifest := manifest.NewMinecraftServer(manifestPath)
	args := automation.ServerArgs{
		StackName:  manifest.Metadata.Name,
		VolumeSize: manifest.Spec.VolumeSize,
		SSH:        manifest.Spec.Ssh,
		Region:     manifest.Spec.Region,
		Size:       manifest.Spec.Size,
	}
	var cloudProvider automation.Automation
	if manifest.Spec.Cloud == "do" {
		common.PrintMixedGreen("Using cloud provider %s\n", "DigitalOcean")
		cloudProvider = cloud.NewDigitalOcean()
	} else if manifest.Spec.Cloud == "civo" {
		common.PrintMixedGreen("Using cloud provider %s\n", "Civo")
		cloudProvider = cloud.NewCivo()
	}

	p := &PulumiProvisioner{
		auto:     cloudProvider,
		manifest: manifest,
		args:     args,
	}
	return p
}
