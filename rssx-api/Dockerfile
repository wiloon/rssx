FROM archlinux:base-20220306.0.49442 AS prod

MAINTAINER wiloon <wiloon.wy@gmail.com>

ARG APP_NAME=rssx-api
# config file
ENV APP_CONFIG_PATH /etc/${APP_NAME}
# bin
COPY ${APP_NAME} /usr/local/bin
COPY config.toml config.toml
COPY config.toml /data/${APP_NAME}/config.toml
COPY rssx-api.db /var/lib/${APP_NAME}/rssx-api.db
ENV APPLICATION_NAME ${APP_NAME}
CMD "/usr/local/bin/${APPLICATION_NAME}"
