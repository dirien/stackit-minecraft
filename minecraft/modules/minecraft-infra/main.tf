terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "1.42.0"
    }
  }
}

resource "openstack_compute_keypair_v2" "minecraft-kp" {
  name       = "${var.minecraft-name}-kp"
  public_key = var.public_key_file
}

resource "openstack_networking_network_v2" "minecraft-net" {
  name           = "${var.minecraft-name}-net"
  admin_state_up = "true"
}


resource "openstack_networking_subnet_v2" "minecraft-snet" {
  name       = "${var.minecraft-name}-snet"
  network_id = openstack_networking_network_v2.minecraft-net.id
  cidr       = var.subnet-cidr
  ip_version = 4
  dns_nameservers = [
    "8.8.8.8",
    "8.8.4.4"]
}

resource "openstack_networking_router_v2" "minecraft-router" {
  name                = "${var.minecraft-name}-router"
  admin_state_up      = "true"
  external_network_id = data.openstack_networking_network_v2.floating.id
}

resource "openstack_networking_router_interface_v2" "minecraft-ri" {
  router_id = openstack_networking_router_v2.minecraft-router.id
  subnet_id = openstack_networking_subnet_v2.minecraft-snet.id
}

resource "openstack_networking_secgroup_v2" "minecraft-sg" {
  name        = "${var.minecraft-name}-sec"
  description = "Security group for the Terraform nodes instances"
}

resource "openstack_networking_secgroup_rule_v2" "minecraft-22-sgr" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.minecraft-sg.id
}

resource "openstack_networking_secgroup_rule_v2" "minecraft-19132-sgr" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = 19132
  port_range_max    = 19132
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.minecraft-sg.id
}


resource "openstack_compute_instance_v2" "minecraft-vm" {
  name        = "${var.minecraft-name}-ubuntu"
  flavor_name = var.flavor
  key_pair    = openstack_compute_keypair_v2.minecraft-kp.name
  security_groups = [
    "default",
    openstack_networking_secgroup_v2.minecraft-sg.name
  ]

  user_data = var.user_data

  network {
    name = openstack_networking_network_v2.minecraft-net.name
  }

  block_device {
    uuid                  = var.ubuntu-image-id
    source_type           = "image"
    boot_index            = 0
    destination_type      = "volume"
    volume_size           = 10
    delete_on_termination = true
  }
}

resource "openstack_blockstorage_volume_v3" "minecraft-vl" {
  name        = "${var.minecraft-name}-vl"
  description = "Storage for the minecraft server"
  size        = 500
  volume_type = "storage_premium_perf2"
}

resource "openstack_compute_volume_attach_v2" "t4ch-volume-attach" {
  volume_id   = openstack_blockstorage_volume_v3.minecraft-vl.id
  instance_id = openstack_compute_instance_v2.minecraft-vm.id
}

resource "openstack_networking_floatingip_v2" "minecraft-fip" {
  pool = var.pool
}

resource "openstack_compute_floatingip_associate_v2" "minecraft-fipa" {
  instance_id = openstack_compute_instance_v2.minecraft-vm.id
  floating_ip = openstack_networking_floatingip_v2.minecraft-fip.address
}

output "minecraft-private" {
  value       = openstack_compute_instance_v2.minecraft-vm.access_ip_v4
  description = "The private ips of the nodes"
}

output "minecraft-public" {
  value       = openstack_networking_floatingip_v2.minecraft-fip.address
  description = "The public ips of the nodes"
}