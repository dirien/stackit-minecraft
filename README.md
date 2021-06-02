# STACKIT Minecraft
STACKIT Bedrock and Java Edition Minecraft Server

Everything done via IaC with Openstack and Terrafrom on STACKIT

## Add Minecraft server

## Add Prometheus

## Add Node Exporter

## Add Ansible for Grafana Cloud remote_write

```bash
cd ansible
ansible-playbook --ask-vault-pass --private-key ../minecraft/ssh/minecraft -i hosts.yaml playbook.yaml
```



