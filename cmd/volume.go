package main

import corev1 "k8s.io/api/core/v1"

// -v /var/lib/lxc/lxcfs/proc/cpuinfo:/proc/cpuinfo:ro
// -v /var/lib/lxc/lxcfs/proc/diskstats:/proc/diskstats:ro
// -v /var/lib/lxc/lxcfs/proc/loadavg:/proc/loadavg:ro
// -v /var/lib/lxc/lxcfs/proc/meminfo:/proc/meminfo:ro
// -v /var/lib/lxc/lxcfs/proc/stat:/proc/stat:ro
// -v /var/lib/lxc/lxcfs/proc/swaps:/proc/swaps:ro
// -v /var/lib/lxc/lxcfs/proc/uptime:/proc/uptime:ro
// -v /var/lib/lxc/lxcfs/sys/devices/system/cpu/online:/sys/devices/system/cpu/online:ro
// -v /var/lib/lxc/:/var/lib/lxc/:ro

const lxcfsVol = "lxcfs"

var volumeMountsTemplate = []corev1.VolumeMount{

	{
		Name:      lxcfsVol,
		MountPath: "/proc/cpuinfo",
		SubPath:   "lxcfs/proc/cpuinfo",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/proc/diskstats",
		SubPath:   "lxcfs/proc/diskstats",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/proc/loadavg",
		SubPath:   "lxcfs/proc/loadavg",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/proc/meminfo",
		SubPath:   "lxcfs/proc/meminfo",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/proc/stat",
		SubPath:   "lxcfs/proc/stat",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/proc/swaps",
		SubPath:   "lxcfs/proc/swaps",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/proc/uptime",
		SubPath:   "lxcfs/proc/uptime",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/sys/devices/system/cpu/online",
		SubPath:   "lxcfs/sys/devices/system/cpu/online",
		ReadOnly:  true,
	},
	{
		Name:      lxcfsVol,
		MountPath: "/var/lib/lxc/",
		ReadOnly:  true,
		MountPropagation: func() *corev1.MountPropagationMode {
			pt := corev1.MountPropagationHostToContainer
			return &pt
		}(),
	},
}

var volumesTemplate = []corev1.Volume{
	{
		Name: lxcfsVol,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxc/",
				Type: func() *corev1.HostPathType {
					pt := corev1.HostPathDirectoryOrCreate
					return &pt
				}(),
			},
		},
	},
}
