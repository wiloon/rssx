#!/bin/sh

cd /home/wiloon/projects/rssx/rssx-ui || exit
yarn install
yarn build
sudo buildah bud -t registry.wiloon.com/rssx-ui:v0.0.1 .
sudo buildah push registry.wiloon.com/rssx-ui:v0.0.1
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman pull registry.wiloon.com/rssx-ui:v0.0.1'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman stop rssx-ui'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman rm rssx-ui'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman run -d --name rssx-ui -p 30090:80 -v /etc/localtime:/etc/localtime:ro -v rssx-ui-logs:/var/log/nginx registry.wiloon.com/rssx-ui:v0.0.1'
ansible -i '192.168.50.100,' all -u root -m shell -a 'podman ps -f name=rssx-ui'
