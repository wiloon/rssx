FROM nginx:1.21.6-alpine AS prod
COPY dist/ /usr/share/nginx/html/
COPY nginx-default.conf /etc/nginx/conf.d/default.conf
