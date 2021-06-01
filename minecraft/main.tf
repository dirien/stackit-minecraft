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

resource "openstack_compute_keypair_v2" "minecraft-kp" {
  name       = "${var.minecraft-name}-kp"
  public_key = file("${var.ssh_key_file}.pub")
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

resource "openstack_networking_secgroup_rule_v2" "minecraft-443-sgr" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = 19132
  port_range_max    = 19132
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.minecraft-sg.id
}

data "template_cloudinit_config" "ubuntu-config" {
  gzip          = true
  base64_encode = true

  # Main cloud-config configuration file.
  part {
    content_type = "text/cloud-config"
    content      = <<-EOF
      #cloud-config
      users:
        - default
        - name: prometheus
          shell: /bin/false
        - name: node_exporter
          shell: /bin/false

      package_update: true

      packages:
        - apt-transport-https
        - ca-certificates
        - curl
        - unzip
        - fail2ban

      fs_setup:
        - label: minecraft
          device: /dev/vdc
          filesystem: xfs
          overwrite: false

      mounts:
        - [/dev/vdc, /minecraft]

      # Enable ipv4 forwarding, required on CIS hardened machines
      write_files:
        - path: /etc/sysctl.d/enabled_ipv4_forwarding.conf
          content: |
            net.ipv4.conf.all.forwarding=1
        - path: /tmp/prometheus.yml
          content: |
            global:
              scrape_interval: 15s

            scrape_configs:
              - job_name: 'prometheus'
                scrape_interval: 5s
                static_configs:
                  - targets: ['localhost:9090']
              - job_name: 'node_exporter'
                scrape_interval: 5s
                static_configs:
                  - targets: ['localhost:9100']
        - path: /etc/systemd/system/prometheus.service
          content: |
            [Unit]
            Description=Prometheus
            Wants=network-online.target
            After=network-online.target

            [Service]
            User=prometheus
            Group=prometheus
            Type=simple
            ExecStart=/usr/local/bin/prometheus \
                --config.file /etc/prometheus/prometheus.yml \
                --storage.tsdb.path /var/lib/prometheus/ \
                --web.console.templates=/etc/prometheus/consoles \
                --web.console.libraries=/etc/prometheus/console_libraries

            [Install]
            WantedBy=multi-user.target
        - path: /etc/systemd/system/node_exporter.service
          content: |
            [Unit]
            Description=Node Exporter
            Wants=network-online.target
            After=network-online.target

            [Service]
            User=node_exporter
            Group=node_exporter
            Type=simple
            ExecStart=/usr/local/bin/node_exporter

            [Install]
            WantedBy=multi-user.target
        - path: /etc/systemd/system/minecraft.service
          content: |
            [Unit]
            Description=STACKIT Minecraft Server
            Documentation=https://www.minecraft.net/en-us/download/server/bedrock

            [Service]
            WorkingDirectory=/minecraft
            Type=simple
            ExecStart=/bin/sh -c "LD_LIBRARY_PATH=. ./bedrock_server"
            Restart=on-failure
            RestartSec=5

            [Install]
            WantedBy=multi-user.target

      runcmd:
        - mkdir /etc/prometheus
        - mkdir /var/lib/prometheus
        - curl -sSL https://github.com/prometheus/prometheus/releases/download/v2.27.1/prometheus-2.27.1.linux-amd64.tar.gz | tar -xz
        - cp prometheus-2.27.1.linux-amd64/prometheus /usr/local/bin/
        - cp prometheus-2.27.1.linux-amd64/promtool /usr/local/bin/
        - chown prometheus:prometheus /usr/local/bin/prometheus
        - chown prometheus:prometheus /usr/local/bin/promtool
        - cp -r prometheus-2.27.1.linux-amd64/consoles /etc/prometheus
        - cp -r prometheus-2.27.1.linux-amd64/console_libraries /etc/prometheus
        - chown -R prometheus:prometheus /var/lib/prometheus
        - chown -R prometheus:prometheus /etc/prometheus/consoles
        - chown -R prometheus:prometheus /etc/prometheus/console_libraries
        - mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml
        - chown prometheus:prometheus /etc/prometheus/prometheus.yml
        - systemctl daemon-reload
        - systemctl start prometheus
        - systemctl enable prometheus

        - curl -sSL https://github.com/prometheus/node_exporter/releases/download/v1.1.2/node_exporter-1.1.2.linux-amd64.tar.gz | tar -xz
        - cp node_exporter-1.1.2.linux-amd64/node_exporter /usr/local/bin
        - chown node_exporter:node_exporter /usr/local/bin/node_exporter
        - systemctl daemon-reload
        - systemctl start node_exporter
        - systemctl enable node_exporter

        - ufw allow ssh
        - ufw allow 5201
        - ufw allow proto udp to 0.0.0.0/0 port 19132
        - echo [DEFAULT] | sudo tee -a /etc/fail2ban/jail.local
        - echo banaction = ufw | sudo tee -a /etc/fail2ban/jail.local
        - echo [sshd] | sudo tee -a /etc/fail2ban/jail.local
        - echo enabled = true | sudo tee -a /etc/fail2ban/jail.local
        - sudo systemctl restart fail2ban
        - curl -sLSf https://minecraft.azureedge.net/bin-linux/bedrock-server-1.16.221.01.zip > /tmp/bedrock-server.zip
        - unzip -o /tmp/bedrock-server.zip -d /minecraft
        - chmod +x /minecraft/bedrock_server
        - sed -ir "s/^[#]*\s*max-players=.*/max-players=100/" /minecraft/server.properties
        - sed -ir "s/^[#]*\s*server-name=.*/server-name=stackit-minecraft/" /minecraft/server.properties
        - sed -ir "s/^[#]*\s*difficulty=.*/difficulty=normal/" /minecraft/server.properties
        - sed -ir "s/^[#]*\s*level-name=.*/level-name=STACKIT/" /minecraft/server.properties
        - sed -ir "s/^[#]*\s*level-seed=.*/level-seed=stackitminecraftrocks/" /minecraft/server.properties
        - systemctl restart minecraft.service
        - systemctl enable minecraft.service
      EOF
  }
}

resource "openstack_compute_instance_v2" "minecraft-vm" {
  name        = "${var.minecraft-name}-ubuntu"
  flavor_name = var.flavor
  key_pair    = openstack_compute_keypair_v2.minecraft-kp.name
  security_groups = [
    "default",
    openstack_networking_secgroup_v2.minecraft-sg.name
  ]

  user_data = data.template_cloudinit_config.ubuntu-config.rendered

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
  name        = "minecraft-vl"
  description = "Storage for the minecraft server"
  size        = 500
  volume_type = "storage_premium_perf2"
}

resource "openstack_compute_volume_attach_v2" "t4ch-volume-attach" {
  volume_id   = openstack_blockstorage_volume_v3.minecraft-vl.id
  instance_id = openstack_compute_instance_v2.minecraft-vm.id
}

resource "openstack_networking_floatingip_v2" "minecraft-fip" {
  pool  = var.pool
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