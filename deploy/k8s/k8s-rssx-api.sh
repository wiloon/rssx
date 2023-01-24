#!/bin/bash

projectName="rssx-api"
cd /home/wiloon/projects/rssx/rssx-api || exit
CGO_ENABLED=1 GOOS=linux GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l /home/wiloon/projects/rssx/rssx-api/rssx-api
# build image
sudo buildah bud -t registry.wiloon.com/rssx-api:v0.0.1 .
# push to registry
sudo buildah push registry.wiloon.com/rssx-api:v0.0.1
echo "y"|sudo podman image prune
rm /home/wiloon/projects/rssx/rssx-api/rssx-api

ansible -i '192.168.50.20,' all -u root -m copy -a 'src=/home/wiloon/workspace/projects/rssx/deploy/k8s/k8s-rssx-api-deployment.yaml dest=/tmp'
ansible -i '192.168.50.20,' all -u root -m shell -a 'kubectl delete -f /tmp/k8s-rssx-api-deployment.yaml'
ansible -i '192.168.50.20,' all -u root -m shell -a 'kubectl create -f /tmp/k8s-rssx-api-deployment.yaml'
