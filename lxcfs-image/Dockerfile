FROM alpine:3

LABEL maintainer="ymping <ympiing@gmail.com>"

ADD . /lxcfs

ARG LXCFS_VERSION
ENV LXCFS_VERSION=${LXCFS_VERSION}

RUN chmod +x /lxcfs/entrypoint.sh && \
    apk add --no-cache dumb-init bash lxcfs=${LXCFS_VERSION}

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/lxcfs/entrypoint.sh"]
