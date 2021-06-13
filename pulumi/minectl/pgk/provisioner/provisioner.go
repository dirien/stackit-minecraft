package provisioner

import (
	"minectl/pgk/automation"
	"minectl/pgk/cloud"
	"minectl/pgk/manifest"
)

type Provisioner interface {
	CreateServer() (*automation.RessourceResults, error)
	DeleteServer(id string) error
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

func (p PulumiProvisioner) DeleteServer(id string) error {
	return p.auto.DeleteServer(id)
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
	cloud := cloud.NewDigitalOcean()
	p := &PulumiProvisioner{
		auto:     cloud,
		manifest: manifest,
		args:     args,
	}
	return p
}
