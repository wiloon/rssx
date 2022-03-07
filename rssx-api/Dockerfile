FROM alpine:3.15.0 AS prod
ARG APP_NAME=rssx-api
COPY ${APP_NAME} /data/${APP_NAME}/${APP_NAME}
COPY config.toml config.toml
COPY config.toml /data/${APP_NAME}/config.toml

ENV APPLICATION_NAME ${APP_NAME}

CMD "/data/${APPLICATION_NAME}/${APPLICATION_NAME}"
