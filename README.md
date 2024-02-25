# Turing Pi v2 K3S Cluster with RK 1 Compute Modules (Project is WIP and not finished!)
This project uses an Ansible playbook to install a K3S 4 node cluster on the RK1 computer modules from the [Turing Pi Project](https://turingpi.com/).
This involves installing 3 servers and 1 agent node.

After the successful K3S installation, Argo CD is installed and linked to this repository using the [App of Apps pattern (cluster bootstrapping)](https://argo-cd.readthedocs.io/en/stable/operator-manual/cluster-bootstrapping/).
This simplifies the handling of the various Helm charts, their configurations and updates.

# Install
## Prerequisites
* The RK1 Ubuntu server image must be flashed to the modules as described in the [Turing Pi Docs - Flashing OS](https://docs.turingpi.com/docs/turing-rk1-flashing-os).
* Passwordless authentication via SSH key must be set up on all nodes.

## Setup K3S on RK1
Copy the hosts-sample.yml and adjust the files. Afterward run the Ansible playbook
```bash
cp hosts-sample.yml hosts.yml
# adjust the ip addresses and nvme settings
# and run the playbook
ansible-playbook -i hosts.yml turingpi.yml
```

TODO ... more Readme and Helm Charts

