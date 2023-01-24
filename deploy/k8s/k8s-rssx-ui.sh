#!/bin/bash

cd /home/wiloon/projects/rssx/rssx-ui || exit
yarn install
yarn build
sudo buildah bud -t registry.wiloon.com/rssx-ui:v0.0.1 .
sudo buildah push registry.wiloon.com/rssx-ui:v0.0.1

ansible -i '192.168.50.20,' all -u root -m copy -a 'src=/home/wiloon/workspace/projects/rssx/deploy/k8s/k8s-rssx-ui-deployment.yaml dest=/tmp'
ansible -i '192.168.50.20,' all -u root -m shell -a 'kubectl delete -f /tmp/k8s-rssx-ui-deployment.yaml'
ansible -i '192.168.50.20,' all -u root -m shell -a 'kubectl create -f /tmp/k8s-rssx-ui-deployment.yaml'
