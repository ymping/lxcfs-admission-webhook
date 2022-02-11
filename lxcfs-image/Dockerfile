FROM alpine:3

LABEL maintainer="ymping <ympiing@gmail.com>"

ADD . /lxcfs

ENV LXCFS_VERSION=4.0.11-r0

RUN chmod +x /lxcfs/entrypoint.sh && \
    apk add --no-cache dumb-init bash lxcfs=${LXCFS_VERSION}

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/lxcfs/entrypoint.sh"]
