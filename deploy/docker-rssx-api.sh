#!/bin/sh
projectName="rssx-api"
cd /home/wiloon/projects/rssx/rssx-api || exit
CGO_ENABLED=1 GOOS=linux GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l /home/wiloon/projects/rssx/rssx-api/rssx-api
sudo buildah bud -t registry.wiloon.com/rssx-api:v0.0.1 .
sudo buildah push registry.wiloon.com/rssx-api:v0.0.1
echo "y"|sudo podman image prune
rm /home/wiloon/projects/rssx/rssx-api/rssx-api

ansible -i '192.168.50.100,' all -u root -m shell -a 'podman pull registry.wiloon.com/rssx-api:v0.0.1'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman stop rssx-api'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman rm rssx-api'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman run -d --name rssx-api -p 3000:8080 -v /etc/localtime:/etc/localtime:ro -v rssx-api-config:/etc/rssx-api -v rssx-api-log:/var/log/rssx-api -v rssx-api-data:/var/lib/rssx-api registry.wiloon.com/rssx-api:v0.0.1'
#ansible -i '192.168.50.100,' all -u root -m shell -a 'podman image prune'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman ps -f name=rssx-api'
