package automation

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type Automation interface {
	CreateServer(args ServerArgs) (*RessourceResults, error)
	DeleteServer(id string) error
	InstallPlugin() error
	Bootstrap(args ServerArgs) pulumi.RunFunc
	UpdateServer(args ServerArgs) (*RessourceResults, error)
}

type ServerArgs struct {
	StackName  string
	Size       string
	Region     string
	SSH        string
	VolumeSize int
}

type RessourceResults struct {
	PublicIP string
}
