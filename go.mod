module github.com/ymping/lxcfs-admission-webhook

go 1.18

replace k8s.io/api => k8s.io/api v0.24.3

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.24.3

replace k8s.io/apimachinery => k8s.io/apimachinery v0.24.4-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.24.3

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.24.3

replace k8s.io/client-go => k8s.io/client-go v0.24.3

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.24.3

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.24.3

replace k8s.io/code-generator => k8s.io/code-generator v0.24.4-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.24.3

replace k8s.io/component-helpers => k8s.io/component-helpers v0.24.3

replace k8s.io/controller-manager => k8s.io/controller-manager v0.24.3

replace k8s.io/cri-api => k8s.io/cri-api v0.25.0-alpha.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.24.3

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.24.3

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.24.3

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.24.3

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.24.3

replace k8s.io/kubectl => k8s.io/kubectl v0.24.3

replace k8s.io/kubelet => k8s.io/kubelet v0.24.3

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.24.3

replace k8s.io/metrics => k8s.io/metrics v0.24.3

replace k8s.io/mount-utils => k8s.io/mount-utils v0.24.4-rc.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.24.3

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.24.3

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.24.3

replace k8s.io/sample-controller => k8s.io/sample-controller v0.24.3

require (
	github.com/golang/glog v1.0.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.24.3
	k8s.io/apimachinery v0.24.3
	k8s.io/kubernetes v1.24.3
)

require (
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiserver v0.24.3 // indirect
	k8s.io/component-base v0.24.3 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
