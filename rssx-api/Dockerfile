FROM ubuntu:22.04 AS prod

MAINTAINER wiloon <wiloon.wy@gmail.com>

ARG APP_NAME=rssx-api
# config file
ENV APP_CONFIG_PATH /etc/${APP_NAME}
# bin
COPY ${APP_NAME} /usr/local/bin
COPY config.toml config.toml
COPY config.toml /data/${APP_NAME}/config.toml
COPY rssx-api.db /var/lib/${APP_NAME}/rssx-api.db
COPY sources.list /etc/apt/sources.list

RUN apt-get -y update && apt-get -y install ca-certificates

ENV APPLICATION_NAME ${APP_NAME}
CMD "/usr/local/bin/${APPLICATION_NAME}"

# COPY test.sh /usr/local/bin
#RUN apt-get update -y
#RUN apt-get install -y socat
#RUN chmod u+x /usr/local/bin/test.sh
#RUN ["chmod", "u+x", "/usr/local/bin/test.sh"]
#RUN ["ls","-l","/usr/local/bin/test.sh"]
#CMD "/usr/local/bin/test.sh"