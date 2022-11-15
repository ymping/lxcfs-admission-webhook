#!/usr/bin/env bash

LXC_PATH="/var/lib/lxc"
LXCFS_PATH="${LXC_PATH}/lxcfs"
LXCFS_SCRIPT_PATH="${LXC_PATH}/script"

# Cleanup
nsenter --target 1 --mount -- fusermount -u $LXCFS_PATH
[[ -d "$LXCFS_PATH" ]] && rm -rf "${LXCFS_PATH:?}"/*

# Prepare
[[ ! -d "$LXCFS_PATH" ]] && mkdir -p $LXCFS_PATH
[[ ! -d "$LXCFS_SCRIPT_PATH" ]] && mkdir -p $LXCFS_SCRIPT_PATH

cat /lxcfs/lxcfs-mount.sh > ${LXCFS_SCRIPT_PATH}/lxcfs-mount.sh
chmod +x ${LXCFS_SCRIPT_PATH}/lxcfs-mount.sh

# Run lxcfs
echo /usr/bin/lxcfs --foreground --enable-loadavg --enable-cfs $LXCFS_PATH
/usr/bin/lxcfs --foreground --enable-loadavg --enable-cfs $LXCFS_PATH
