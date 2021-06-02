terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.61.0"
    }
  }
  backend "azurerm" {
    storage_account_name = "aebi"
    container_name       = "stackit-minecraft-server-state"
    key                  = "stackit-minecraft-server.tfstate"
  }
}

module "minecraft-bedrock-server" {
  source          = "./modules/minecraft-infra"
  minecraft-name  = "minecraft-bedrock"
  public_key_file = file("${var.ssh_key_file}.pub")
  subnet-cidr     = "10.1.10.0/24"
  user_data       = data.template_cloudinit_config.bedrock-config.rendered
}

output "minecraft-public" {
  value = module.minecraft-bedrock-server.minecraft-public
}