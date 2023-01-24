FROM archlinux:base-20230108.0.116909 AS prod

MAINTAINER wiloon <wiloon.wy@gmail.com>

ARG APP_NAME=rssx-api
# config file
ENV APP_CONFIG_PATH /etc/${APP_NAME}
# bin
COPY ${APP_NAME} /usr/local/bin
COPY config.toml config.toml
COPY config.toml /data/${APP_NAME}/config.toml
ENV APPLICATION_NAME ${APP_NAME}
CMD "/usr/local/bin/${APPLICATION_NAME}"
