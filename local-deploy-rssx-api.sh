#!/bin/sh
cd /home/wiloon/projects/rssx/rssx-api || exit
CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.io go build -a -o rssx-api main.go
ls -l /home/wiloon/projects/rssx/rssx-api/rssx-api
sudo buildah bud -t registry.wiloon.com/rssx-api:v0.0.1 .
sudo buildah push registry.wiloon.com/rssx-api:v0.0.1
rm /home/wiloon/projects/rssx/rssx-api/rssx-api
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman pull registry.wiloon.com/rssx-api:v0.0.1'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman stop rssx-api'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman rm rssx-api'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman run -d --name rssx-api -p 3000:8080 -v /etc/localtime:/etc/localtime:ro -v rssx-logs:/data/rssx/logs registry.wiloon.com/rssx-api:v0.0.1'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman ps -f name=rssx-api'