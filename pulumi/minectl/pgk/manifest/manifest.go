package manifest

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"log"
	"sigs.k8s.io/yaml"
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

//go:embed schema.json
var schema string

func validate(manifest []byte) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	yaml, err := yaml.YAMLToJSON(manifest)
	if err != nil {
		log.Fatal(err)
	}
	documentLoader := gojsonschema.NewStringLoader(string(yaml))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Fatal(err)
	}

	if !result.Valid() {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
			return errors.New("validation error")
		}
	}
	return nil
}

func NewMinecraftServer(manifestPath string) *MinecraftServer {

	var server MinecraftServer
	manifestFile, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}
	err = validate(manifestFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(manifestFile, &server)
	if err != nil {
		panic(err)
	}
	return &server
}
