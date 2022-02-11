#!/usr/bin/env bash

# adjust fuse.lxcfs filesystem mount in container after/before LXCFS DaemonSet start/stop

PATH=$PATH:/bin
LXC_PATH="/var/lib/lxc"
LXCFS_PATH="${LXC_PATH}/lxcfs"

UMOUNT=false
REMOUNT=false

# echo script usage
usage() {
  cat <<EOF

Adjust the fuse.lxcfs filesystem mount in container

  - /var/lib/lxc/lxcfs/proc/cpuinfo:/proc/cpuinfo
  - /var/lib/lxc/lxcfs/proc/diskstats:/proc/diskstats
  - /var/lib/lxc/lxcfs/proc/meminfo:/proc/meminfo
  - /var/lib/lxc/lxcfs/proc/stat:/proc/stat
  - /var/lib/lxc/lxcfs/proc/swaps:/proc/swaps
  - /var/lib/lxc/lxcfs/proc/uptime:/proc/uptime
  - /var/lib/lxc/lxcfs/proc/loadavg:/proc/loadavg
  - /var/lib/lxc/lxcfs/sys/devices/system/cpu/online:/sys/devices/system/cpu/online

Umount all fuse.lxcfs filesystem mount in container before LXCFS daemonset pod stop
or fix "Transport endpoint is not connected" error that case by LXCFS pod unexpected stop
by remount fuse.lxcfs filesystem after LXCFS daemonset pod start

usage: ${0} [OPTIONS]

The following one flag are required

  --umount            umount fuse.lxcfs filesystem mount in container
  --remount           umount fuse.lxcfs filesystem mount in container and remount it

EOF

  exit 0
}

# check python3, nsenter, crictl or docker command exist on k8s cluster
pre_check() {
  if ! command -v python3 >/dev/null; then
    echo python3 interpreter not found on host, exit
    exit 1
  fi

  if ! command -v nsenter >/dev/null; then
    echo nsenter not found on host, exit
    exit 1
  fi

  if ! command -v crictl >/dev/null && ! command -v docker >/dev/null; then
    echo container cli command crictl or docker not found on host, exit
    exit 1
  fi
}

# remount fuse.lxcfs filesystem in container
# if fuse.lxcfs mount point is broken in container, umount and mount it again
# if mount point is ok and fuse.lxcfs filesystem mount in container, mount it again
lxcfs_remount() {
  container_pid=$1

  # mount proc
  for file in cpuinfo diskstats loadavg meminfo stat swaps uptime; do
    # /proc/$file mount point exist as type of fuse.lxcfs in mount table and not connected, umount it
    if ! nsenter -t "$container_pid" -m -- test -e "/proc/$file" && nsenter -t "$container_pid" -m -p -- mount -t fuse.lxcfs | grep -qs "/proc/$file"; then
      echo nsenter -t "$container_pid" -m -p -- umount -v "/proc/$file"
      nsenter -t "$container_pid" -m -p -- umount -v "/proc/$file"
    fi
    # /proc/$file mount point not in mount table and fuse.lxcfs filesystem mount in container, remount it
    if nsenter -t "$container_pid" -m -- test -e "$LXCFS_PATH/proc/$file" && ! nsenter -t "$container_pid" -m -p -- mount -t fuse.lxcfs | grep -qs "/proc/$file"; then
      echo nsenter -t "$container_pid" -m -- mount -B -v -o ro "$LXCFS_PATH/proc/$file" "/proc/$file"
      nsenter -t "$container_pid" -m -- mount -B -v -o ro "$LXCFS_PATH/proc/$file" "/proc/$file"
    fi
  done

  # mount online cpu
  if ! nsenter -t "$container_pid" -m -- test -e "/sys/devices/system/cpu/online" && nsenter -t "$container_pid" -m -p -- mount -t fuse.lxcfs | grep -qs "/sys/devices/system/cpu/online"; then
    echo nsenter -t "$container_pid" -m -p -- umount -v "/sys/devices/system/cpu/online"
    nsenter -t "$container_pid" -m -p -- umount -v "/sys/devices/system/cpu/online"
  fi
  if nsenter -t "$container_pid" -m -- test -e "$LXCFS_PATH/sys/devices/system/cpu/online" && ! nsenter -t "$container_pid" -m -p -- mount -t fuse.lxcfs | grep -qs "/sys/devices/system/cpu/online"; then
    echo nsenter -t "$container_pid" -m -- mount -B -v -o ro "$LXCFS_PATH/sys/devices/system/cpu/online" "/sys/devices/system/cpu/online"
    nsenter -t "$container_pid" -m -- mount -B -v -o ro "$LXCFS_PATH/sys/devices/system/cpu/online" "/sys/devices/system/cpu/online"
  fi
}

# umount fuse.lxcfs filesystem in container
lxcfs_umount() {
  container_pid=$1

  # umount proc
  for file in cpuinfo diskstats loadavg meminfo stat swaps uptime; do
    if nsenter -t "$container_pid" -m -p -- mount -t fuse.lxcfs | grep -qs "/proc/$file"; then
      echo nsenter -t "$container_pid" -m -p -- umount -v "/proc/$file"
      nsenter -t "$container_pid" -m -p -- umount -v "/proc/$file"
    fi
  done

  # umount online cpu
  if nsenter -t "$container_pid" -m -p -- mount -t fuse.lxcfs | grep -qs "/sys/devices/system/cpu/online"; then
    echo nsenter -t "$container_pid" -m -p -- umount -v "/sys/devices/system/cpu/online"
    nsenter -t "$container_pid" -m -p -- umount -v "/sys/devices/system/cpu/online"
  fi
}

# call lxcfs_remount or lxcfs_umount by shell args
lxcfs_mount() {
  container_pid=$1

  if [[ "$REMOUNT" == true ]]; then
    lxcfs_remount "$container_pid"
  elif [[ "$UMOUNT" == true ]]; then
    lxcfs_umount "$container_pid"
  else
    echo "unknown action got, exit"
    exit 1
  fi
}

docker_cli() {
  # skip pause container
  # skip container in namespaces kube-system or kube-public
  containers=$(docker ps | grep -v -E "pause|kube-system|kube-public" | awk 'NR > 1 {print $1}')

  for container in $containers; do
    mount_point=$(docker inspect --format "{{ range .Mounts }}{{ if eq .Destination \"$LXC_PATH\"  }}{{ .Source }}{{ end }}{{ end }}" "$container")

    if [[ "$mount_point" == "$LXC_PATH" ]]; then
      # skip itself, the lxcfs daemonset container
      # check by has environment LXCFS_VERSION=xxx set at Dockerfile
      container_envs=$(docker inspect --format '{{ range .Config.Env }} {{ . }} {{ end }}' "$container")
      if [[ "${container_envs[*]}" =~ "LXCFS_VERSION" ]]; then
        break
      fi
      pid=$(docker inspect --format '{{.State.Pid}}' "$container")
      echo "adjust lxcfs mount in container $container on $(hostname) with docker"
      lxcfs_mount "$pid"
    fi
  done
}

crictl_cli() {
  # prepare crictl connect endpoint
  # https://github.com/kubernetes-sigs/cri-tools/blob/33d7f05e2ad599eb269d468447615b5380191a69/docs/crictl.md?plain=1#L99
  endpoints=("/var/run/dockershim.sock" "/run/containerd/containerd.sock" "/run/crio/crio.sock" "/var/run/cri-dockerd.sock")
  for endpoint in "${endpoints[@]}"; do
    if [[ -S $endpoint ]]; then
      export CONTAINER_RUNTIME_ENDPOINT=$endpoint
      break
    fi
  done

  # get pod's id that pod annotations with "mutating.lxcfs-admission-webhook.io/status: mutated"
  # and without lable "app: lxcfs-ds"
  pods=$(crictl pods --output table --state ready --verbose |
    awk -v RS= '/mutating.lxcfs-admission-webhook.io\/status -> mutated/ && ! /app -> lxcfs-ds/ {print $0"\n"}' |
    awk -F ": " '/^ID:/ {print $2}')

  for pod in $pods; do
    containers=$(crictl ps --quiet --pod "$pod")
    for container in $containers; do
      pid=$(crictl inspect --output=json "$container" | python3 -c "import sys, json; print(json.load(sys.stdin)['info']['pid'])")
      echo "adjust lxcfs mount in container $container on $(hostname) with crictl"
      lxcfs_mount "$pid"
    done
  done
}

main() {
  pre_check

  if [[ $# -eq 1 ]]; then
    case $1 in
    --umount)
      UMOUNT=true
      ;;
    --remount)
      REMOUNT=true
      # wait 3 seconds to start lxcfs
      # for post-start hook is executed immediately after a container is created
      sleep 3
      ;;
    *)
      usage
      ;;
    esac
  else
    usage
  fi

  if command -v docker >/dev/null; then
    docker_cli
  else
    crictl_cli
  fi
}

main "$@"
