package main

import (
	_ "embed"
	"log"
	"minectl/pgk/provisioner"
)

func main() {
	log.Println("hello")
	var stackName = "minecraft-server"
	do := provisioner.NewProvisioner("/Users/dirien/Tools/repos/stackit-minecraft/pulumi/minectl/cmd/minectl/server.yaml")
	var create = false
	var update = true

	if create {
		res, err := do.CreateServer()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(res.PublicIP)
	} else {
		if update {
			do.UpdateServer()
		} else {
			do.DeleteServer(stackName)
		}
	}
}
