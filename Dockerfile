FROM golang:1.17 AS build
ENV GO111MODULE on
WORKDIR /workdir
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.io go build -a rssx.go

FROM alpine AS prod
COPY --from=build /workdir/rssx /data/rssx/
COPY config.toml config.toml
COPY config.toml /data/rssx/config.toml
CMD ["/data/rssx/rssx"]
