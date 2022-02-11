FROM golang:1.17-alpine3.15 as build

WORKDIR /build

ADD . .

RUN export GOPROXY=https://goproxy.io,direct &&\
    go build -o lxcfs-admission-webhook ./cmd/

FROM alpine:3.15

LABEL maintainer="ymping <ympiing@gmail.com>"

WORKDIR /lxcfs

COPY --from=build /build/lxcfs-admission-webhook /lxcfs/lxcfs-admission-webhook

EXPOSE 8443

ENTRYPOINT ["/lxcfs/lxcfs-admission-webhook"]
