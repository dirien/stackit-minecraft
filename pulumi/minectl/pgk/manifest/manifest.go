package manifest

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type MinecraftServer struct {
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Spec struct {
	Cloud      string `yaml:"cloud"`
	Ssh        string `yaml:"ssh"`
	Region     string `yaml:"region"`
	Size       string `yaml:"size"`
	VolumeSize int    `yaml:"volumeSize"`
}

func NewMinecraftServer(manifestPath string) *MinecraftServer {
	var server MinecraftServer
	manifestFile, err := ioutil.ReadFile(manifestPath)
	err = yaml.Unmarshal(manifestFile, &server)
	if err != nil {
		panic(err)
	}

	fmt.Print(server)
	return &server
}
