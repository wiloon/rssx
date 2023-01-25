FROM alpine:3.17.1 AS prod

MAINTAINER wiloon <wiloon.wy@gmail.com>

ARG APP_NAME=rssx-api
# config file
ENV APP_CONFIG_PATH /etc/${APP_NAME}
# bin
COPY ${APP_NAME} /usr/local/bin
COPY config.toml config.toml
COPY config.toml /data/${APP_NAME}/config.toml
COPY rssx-api.db /var/lib/${APP_NAME}/rssx-api.db

COPY test.sh /usr/local/bin

ENV APPLICATION_NAME ${APP_NAME}
#CMD "/usr/local/bin/${APPLICATION_NAME}"
CMD "/usr/local/bin/test.sh"
