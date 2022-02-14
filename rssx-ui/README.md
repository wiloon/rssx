# rssx-web

### Compiles and hot-reloads for development
```
yarn serve
```

### Compiles and minifies for production
```
yarn build
```

### Run your unit tests
```
yarn test:unit
```

### Run your end-to-end tests
```
yarn test:e2e
```

### Lints and fixes files
```
yarn lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

buildah bud -f Dockerfile -t rssx .

### deploy
```bash
podman stop rssx && podman rm rssx && podman run -d --name rssx -p 30090:80/tcp -v /etc/localtime:/etc/localtime:ro -v rssx-web-data:/var/log/nginx registry.cn-qingdao.aliyuncs.com/pingd/rssx:1.0.7 && podman ps

```
