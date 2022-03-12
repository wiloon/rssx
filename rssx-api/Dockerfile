FROM alpine:3.15.0 AS prod

RUN apk update
RUN apk upgrade
RUN apk add --update go=1.17.8 gcc=6.3.0-r4 g++=6.3.0-r4

ARG APP_NAME=rssx-api
COPY ${APP_NAME} /data/${APP_NAME}/${APP_NAME}
COPY config.toml config.toml
COPY config.toml /data/${APP_NAME}/config.toml

ENV APPLICATION_NAME ${APP_NAME}

CMD "/data/${APPLICATION_NAME}/${APPLICATION_NAME}"
