variable "pool" {
  default = "floating-net"
}

variable "public_key_file" {}

variable "minecraft-name" {}

variable "flavor" {
  default = "c1.3"
}

variable "subnet-cidr" {
  default = "10.1.10.0/24"
}

variable "ubuntu-image-id" {
  default = "b017f5da-86e2-49ec-98ce-14250f758bfa"
}

variable "user_data" {}