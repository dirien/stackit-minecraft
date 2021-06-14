package cloud

import (
	_ "embed"
)

//go:embed cloud-config.yaml
var cloudConfig string
