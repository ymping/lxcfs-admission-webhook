package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubernetes/pkg/apis/core/v1"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

var validMutatingKindList = []metav1.GroupVersionKind{
	{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	},
}

var validMutatingOperationList = []admissionv1.Operation{
	admissionv1.Create,
}

const (
	admissionWebhookAnnotationEnableKey = "mutating.lxcfs-admission-webhook.io/enable"
	admissionWebhookAnnotationStatusKey = "mutating.lxcfs-admission-webhook.io/status"

	admissionWebhookSuccessFlag  = "mutated"
	admissionWebhookConflictFlag = "conflict"
	admissionWebhookSkipFlag     = "skip"

	admissionWebhookResponseAPIVersion = "admission.k8s.io/v1"
	admissionWebhookResponseKind       = "AdmissionReview"
)

// WebhookServer lxcfs admission webhook server
type WebhookServer struct {
	server *http.Server
}

// WhSvrParameters webhook server parameters
type WhSvrParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// (https://github.com/kubernetes/kubernetes/issues/57982)
func applyDefaultsWorkaround(volumes []corev1.Volume) {
	defaulter.Default(&corev1.Pod{
		Spec: corev1.PodSpec{
			Volumes: volumes,
		},
	})
}

// Check whether the target resoured need to be mutated
func mutationRequired(ignoredNSList []string, validKindList []metav1.GroupVersionKind, validOperationList []admissionv1.Operation, admissionReview *admissionv1.AdmissionReview) bool {
	admissionRequest := admissionReview.Request

	var pod corev1.Pod
	if err := json.Unmarshal(admissionRequest.Object.Raw, &pod); err != nil {
		return false
	}

	// skip special kubernete system namespaces
	for _, namespace := range ignoredNSList {
		if admissionRequest.Namespace == namespace {
			glog.Infof("Skip mutation for %v for it's in special namespace: %v", pod.GenerateName, admissionRequest.Namespace)
			return false
		}
	}

	// verify operation
	validOp := false
	for _, operation := range validOperationList {
		if admissionRequest.Operation == operation {
			validOp = true
		}
	}
	if !validOp {
		return false
	}

	// verify the kind got
	// in case MutatingWebhookConfiguration rules capture wrong kind
	validKind := false
	for _, kind := range validKindList {
		if admissionRequest.Kind == kind {
			validKind = true
		}
	}
	if !validKind {
		return false
	}

	annotations := pod.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	status := annotations[admissionWebhookAnnotationStatusKey]

	// determine whether to perform mutation based on annotation for the target resource
	var required bool
	if strings.ToLower(status) == admissionWebhookSuccessFlag {
		required = false
	} else {
		switch strings.ToLower(annotations[admissionWebhookAnnotationEnableKey]) {
		default:
			required = true
		case "n", "no", "false", "off":
			required = false
		}
	}

	glog.Infof("Mutation policy for %v/%v: status: %q required:%v", admissionRequest.Namespace, pod.GenerateName, status, required)
	return required
}

// volumeMountConflictCheck check VolumeMount of target and added has same Name or MountPath
func volumeMountConflictCheck(target, added []corev1.VolumeMount) bool {
	for _, origin := range target {
		for _, add := range added {
			if origin.Name == add.Name || origin.MountPath == add.MountPath {
				return true
			}
		}
	}
	return false
}

// volumeConflictCheck check Volume of target and added has same Name
func volumeConflictCheck(target, added []corev1.Volume) bool {
	for _, origin := range target {
		for _, add := range added {
			if origin.Name == add.Name {
				return true
			}
		}
	}
	return false
}

func patchVolumeMount(target, added []corev1.VolumeMount, targetIndex int) (patches []patchOperation) {
	if len(target) == 0 {
		path := fmt.Sprintf("/spec/containers/%d/volumeMounts", targetIndex)
		op := patchOperation{
			Op:    "add",
			Path:  path,
			Value: added,
		}
		patches = append(patches, op)
	} else {
		path := fmt.Sprintf("/spec/containers/%d/volumeMounts/-", targetIndex)
		for _, volumeMount := range added {
			op := patchOperation{
				Op:    "add",
				Path:  path,
				Value: volumeMount,
			}
			patches = append(patches, op)
		}
	}
	return patches
}

func patchVolume(target, added []corev1.Volume) (patches []patchOperation) {
	if len(target) == 0 {
		op := patchOperation{
			Op:    "add",
			Path:  "/spec/volumes",
			Value: added,
		}
		patches = append(patches, op)
	} else {
		for _, volume := range added {
			op := patchOperation{
				Op:    "add",
				Path:  "/spec/volumes/-",
				Value: volume,
			}
			patches = append(patches, op)
		}
	}
	return patches
}

func patchAnnotation(target, added map[string]string) (patches []patchOperation) {
	for key, value := range added {
		var op = patchOperation{
			Op:   "add",
			Path: "/metadata/annotations",
			Value: map[string]string{
				key: value,
			},
		}

		if target != nil {
			if _, ok := target[key]; ok {
				op.Op = "replace"
				op.Path = "/metadata/annotations/" + escapeJSONPointerValue(key)
				op.Value = value
			}
		}

		patches = append(patches, op)
	}
	return patches
}

func escapeJSONPointerValue(in string) string {
	step := strings.Replace(in, "~", "~0", -1)
	return strings.Replace(step, "/", "~1", -1)
}

func patchConflictCheck(pod *corev1.Pod, volumesTemplate []corev1.Volume, volumeMountsTemplate []corev1.VolumeMount) bool {
	containers := pod.Spec.Containers
	for _, container := range containers {
		if volumeMountConflictCheck(container.VolumeMounts, volumeMountsTemplate) {
			return true
		}
	}

	return volumeConflictCheck(pod.Spec.Volumes, volumesTemplate)
}

// create mutation patch for resoures
func createPatch(pod *corev1.Pod, volumesTemplate []corev1.Volume, volumeMountsTemplate []corev1.VolumeMount, annotations map[string]string) ([]byte, error) {
	var patches []patchOperation

	containers := pod.Spec.Containers
	for idx, container := range containers {
		patches = append(patches, patchVolumeMount(container.VolumeMounts, volumeMountsTemplate, idx)...)
	}
	patches = append(patches, patchVolume(pod.Spec.Volumes, volumesTemplate)...)
	patches = append(patches, patchAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patches)
}

// main mutation process
func (whsvr *WebhookServer) mutate(admissionReview *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	admissionRequest := admissionReview.Request

	var pod corev1.Pod
	if err := json.Unmarshal(admissionRequest.Object.Raw, &pod); err != nil {
		glog.Errorf("Could not unmarshal raw object: %v", err)
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		admissionRequest.Kind, admissionRequest.Namespace, admissionRequest.Name, pod.GenerateName, admissionRequest.UID, admissionRequest.Operation, admissionRequest.UserInfo)

	// Workaround: https://github.com/kubernetes/kubernetes/issues/57982
	applyDefaultsWorkaround(volumesTemplate)
	var annotations = make(map[string]string)
	var volumesTemplateToPatch []corev1.Volume
	var volumeMountsTemplateToPatch []corev1.VolumeMount

	if !mutationRequired(ignoredNamespaces, validMutatingKindList, validMutatingOperationList, admissionReview) {
		glog.Infof("Skipping mutation for %s/%s, UID=%s due to policy check", admissionRequest.Namespace, pod.GenerateName, admissionRequest.UID)
		annotations[admissionWebhookAnnotationStatusKey] = admissionWebhookSkipFlag
	} else if patchConflictCheck(&pod, volumesTemplate, volumeMountsTemplate) {
		glog.Infof("Skipping mutation for %s/%s, UID=%s due to volume or volume mount conflict", admissionRequest.Namespace, pod.GenerateName, admissionRequest.UID)
		annotations[admissionWebhookAnnotationStatusKey] = admissionWebhookConflictFlag
	} else {
		annotations[admissionWebhookAnnotationStatusKey] = admissionWebhookSuccessFlag
		volumesTemplateToPatch = volumesTemplate
		volumeMountsTemplateToPatch = volumeMountsTemplate
	}

	patchBytes, err := createPatch(&pod, volumesTemplateToPatch, volumeMountsTemplateToPatch, annotations)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// serve method for webhook server
func (whsvr *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *admissionv1.AdmissionResponse
	ar := admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else if ar.Request == nil {
		glog.Error("Got nil admissionRequest object after deserializer http request body")
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: "Got nil admissionRequest object after deserializer http request body",
			},
		}
	} else {
		admissionResponse = whsvr.mutate(&ar)
	}

	admissionReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       admissionWebhookResponseKind,
			APIVersion: admissionWebhookResponseAPIVersion,
		},
	}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (whsvr *WebhookServer) ping(w http.ResponseWriter, _ *http.Request) {
	if _, err := fmt.Fprintf(w, "pong"); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
