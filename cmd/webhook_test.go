package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"gotest.tools/assert"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func NewWebhookServer() *WebhookServer {
	return &WebhookServer{
		server: &http.Server{
			Addr: fmt.Sprintf(":%v", 8080),
		},
	}
}

func GetAdmissionReviewExample() *admissionv1.AdmissionReview {
	example := []byte(`{
    "kind": "AdmissionReview",
    "apiVersion": "admission.k8s.io/v1",
    "request": {
        "uid": "3c00fd3b-a64b-4120-9b75-1d49ddb95774",
        "kind": {
            "group": "",
            "version": "v1",
            "kind": "Pod"
        },
        "resource": {
            "group": "",
            "version": "v1",
            "resource": "pods"
        },
        "requestKind": {
            "group": "",
            "version": "v1",
            "kind": "Pod"
        },
        "requestResource": {
            "group": "",
            "version": "v1",
            "resource": "pods"
        },
        "namespace": "demo2",
        "operation": "CREATE",
        "userInfo": {
            "username": "system:serviceaccount:kube-system:replicaset-controller",
            "uid": "3eca5dd9-db5c-4ce5-83a3-737f1ef331ea",
            "groups": [
                "system:serviceaccounts",
                "system:serviceaccounts:kube-system",
                "system:authenticated"
            ]
        },
        "object": {
            "kind": "Pod",
            "apiVersion": "v1",
            "metadata": {
                "generateName": "nginx-6fc77dcb7c-",
                "creationTimestamp": null,
                "labels": {
                    "app": "nginx",
                    "pod-template-hash": "6fc77dcb7c"
                },
                "ownerReferences": [
                    {
                        "apiVersion": "apps/v1",
                        "kind": "ReplicaSet",
                        "name": "nginx-6fc77dcb7c",
                        "uid": "932f1aea-0723-4eb1-ba49-76e36fcd7a5d",
                        "controller": true,
                        "blockOwnerDeletion": true
                    }
                ],
                "managedFields": [
                    {
                        "manager": "kube-controller-manager",
                        "operation": "Update",
                        "apiVersion": "v1",
                        "time": "2021-12-12T15:26:53Z",
                        "fieldsType": "FieldsV1",
                        "fieldsV1": {
                            "f:metadata": {
                                "f:generateName": {},
                                "f:labels": {
                                    ".": {},
                                    "f:app": {},
                                    "f:pod-template-hash": {}
                                },
                                "f:ownerReferences": {
                                    ".": {},
                                    "k:{\"uid\":\"932f1aea-0723-4eb1-ba49-76e36fcd7a5d\"}": {
                                        ".": {},
                                        "f:apiVersion": {},
                                        "f:blockOwnerDeletion": {},
                                        "f:controller": {},
                                        "f:kind": {},
                                        "f:name": {},
                                        "f:uid": {}
                                    }
                                }
                            },
                            "f:spec": {
                                "f:containers": {
                                    "k:{\"name\":\"nginx\"}": {
                                        ".": {},
                                        "f:image": {},
                                        "f:imagePullPolicy": {},
                                        "f:name": {},
                                        "f:ports": {
                                            ".": {},
                                            "k:{\"containerPort\":80,\"protocol\":\"TCP\"}": {
                                                ".": {},
                                                "f:containerPort": {},
                                                "f:protocol": {}
                                            }
                                        },
                                        "f:resources": {},
                                        "f:terminationMessagePath": {},
                                        "f:terminationMessagePolicy": {}
                                    }
                                },
                                "f:dnsPolicy": {},
                                "f:enableServiceLinks": {},
                                "f:restartPolicy": {},
                                "f:schedulerName": {},
                                "f:securityContext": {},
                                "f:terminationGracePeriodSeconds": {}
                            }
                        }
                    }
                ]
            },
            "spec": {
                "volumes": [
                    {
                        "name": "default-token-46sr4",
                        "secret": {
                            "secretName": "default-token-46sr4"
                        }
                    }
                ],
                "containers": [
                    {
                        "name": "nginx",
                        "image": "nginx:1.21",
                        "ports": [
                            {
                                "containerPort": 80,
                                "protocol": "TCP"
                            }
                        ],
                        "resources": {},
                        "volumeMounts": [
                            {
                                "name": "default-token-46sr4",
                                "readOnly": true,
                                "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
                            }
                        ],
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File",
                        "imagePullPolicy": "IfNotPresent"
                    }
                ],
                "restartPolicy": "Always",
                "terminationGracePeriodSeconds": 30,
                "dnsPolicy": "ClusterFirst",
                "serviceAccountName": "default",
                "serviceAccount": "default",
                "securityContext": {},
                "schedulerName": "default-scheduler",
                "tolerations": [
                    {
                        "key": "node.kubernetes.io/not-ready",
                        "operator": "Exists",
                        "effect": "NoExecute",
                        "tolerationSeconds": 300
                    },
                    {
                        "key": "node.kubernetes.io/unreachable",
                        "operator": "Exists",
                        "effect": "NoExecute",
                        "tolerationSeconds": 300
                    }
                ],
                "priority": 0,
                "dnsConfig": {
                    "options": [
                        {
                            "name": "single-request-reopen",
                            "value": ""
                        },
                        {
                            "name": "timeout",
                            "value": "2"
                        }
                    ]
                },
                "enableServiceLinks": true,
                "preemptionPolicy": "PreemptLowerPriority"
            },
            "status": {}
        },
        "oldObject": null,
        "dryRun": false,
        "options": {
            "kind": "CreateOptions",
            "apiVersion": "meta.k8s.io/v1"
        }
    }
}`)

	admissionReviewExample := admissionv1.AdmissionReview{}
	if _, _, err := deserializer.Decode(example, nil, &admissionReviewExample); err != nil {
		glog.Errorf("Can't decode request body: %v", err)
	}

	return &admissionReviewExample
}

func TestWebhookServerPing(t *testing.T) {
	whsvr := NewWebhookServer()

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(whsvr.ping)

	handler.ServeHTTP(rr, req)

	expectedBody := "pong"

	assert.Equal(t, rr.Code, http.StatusOK, fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK))
	assert.Equal(t, rr.Body.String(), expectedBody, fmt.Sprintf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody))
}

func TestWebhookServerServe(t *testing.T) {
	whsvr := NewWebhookServer()

	testCases := []struct {
		name     string
		method   string
		header   map[string][]string
		reqBody  string
		httpCode int
		respBody string
	}{
		{"test empty body", http.MethodPost, make(http.Header), "", http.StatusBadRequest, "empty body\n"},
		{"test content type", http.MethodPost, http.Header{"Content-Type": {"text/html"}}, "{}", http.StatusUnsupportedMediaType, "invalid Content-Type, expect `application/json`\n"},
		{"test decode request", http.MethodPost, http.Header{"Content-Type": {"application/json"}}, "{}", http.StatusOK, "Got nil admissionRequest object after deserializer http request body"},
	}

	for _, tc := range testCases {
		t.Logf("Test case for: %s", tc.name)
		req, err := http.NewRequest(tc.method, "/mutate", strings.NewReader(tc.reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header = tc.header

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(whsvr.serve)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, tc.httpCode, fmt.Sprintf("handler returned wrong status code: got %v want %v", rr.Code, tc.httpCode))
		assert.Equal(t, strings.Contains(rr.Body.String(), tc.respBody), true, fmt.Sprintf("handler returned unexpected body: got %v not contain %v", rr.Body.String(), tc.respBody))
	}
}

func TestWebhookServerMutate(t *testing.T) {
	whsvr := NewWebhookServer()

	exceptSkipJsonPatch := "{\"op\":\"add\",\"path\":\"/metadata/annotations\",\"value\":{\"mutating.lxcfs-admission-webhook.io/status\":\"skip\"}}"
	exceptReplaceSkipJsonPatch := "{\"op\":\"replace\",\"path\":\"/metadata/annotations/mutating.lxcfs-admission-webhook.io~1status\",\"value\":\"skip\"}"
	exceptMutatedJsonPatch := "{\"op\":\"add\",\"path\":\"/metadata/annotations\",\"value\":{\"mutating.lxcfs-admission-webhook.io/status\":\"mutated\"}}"
	exceptConflictJsonPatch := "{\"op\":\"add\",\"path\":\"/metadata/annotations\",\"value\":{\"mutating.lxcfs-admission-webhook.io/status\":\"conflict\"}}"

	admissionReviewExample := GetAdmissionReviewExample()
	admissionReviewExampleRequest, _ := json.Marshal(admissionReviewExample)

	admissionReviewWithNamespaceSystem := admissionv1.AdmissionReview{}
	admissionReviewExample.DeepCopyInto(&admissionReviewWithNamespaceSystem)
	admissionReviewWithNamespaceSystem.Request.Namespace = metav1.NamespaceSystem
	admissionReviewWithNamespaceSystemRequest, _ := json.Marshal(admissionReviewWithNamespaceSystem)

	admissionReviewWithScaleKind := admissionv1.AdmissionReview{}
	admissionReviewExample.DeepCopyInto(&admissionReviewWithScaleKind)
	admissionReviewWithScaleKind.Request.Kind = metav1.GroupVersionKind{
		Group:   "autoscaling",
		Version: "v1",
		Kind:    "Scale",
	}
	admissionReviewWithScaleKindRequest, _ := json.Marshal(admissionReviewWithScaleKind)

	admissionReviewWithDenyAnnotation := admissionv1.AdmissionReview{}
	admissionReviewExample.DeepCopyInto(&admissionReviewWithDenyAnnotation)
	var podWithDenyAnnotation corev1.Pod
	if err := json.Unmarshal(admissionReviewWithDenyAnnotation.Request.Object.Raw, &podWithDenyAnnotation); err != nil {
		t.Error(err)
	}
	podWithDenyAnnotation.SetAnnotations(map[string]string{
		admissionWebhookAnnotationEnableKey: "No",
		admissionWebhookAnnotationStatusKey: "test",
	})
	podWithDenyAnnotationRaw, _ := json.Marshal(podWithDenyAnnotation)
	admissionReviewWithDenyAnnotation.Request.Object.Raw = podWithDenyAnnotationRaw
	admissionReviewWithDenyAnnotationRequest, _ := json.Marshal(admissionReviewWithDenyAnnotation)

	admissionReviewWithVolumeMountConflict := admissionv1.AdmissionReview{}
	admissionReviewExample.DeepCopyInto(&admissionReviewWithVolumeMountConflict)
	var podWithVolumeMountConflict corev1.Pod
	if err := json.Unmarshal(admissionReviewWithVolumeMountConflict.Request.Object.Raw, &podWithVolumeMountConflict); err != nil {
		t.Error(err)
	}
	volumeMounts := volumeMountsTemplate[0:1]
	for i, container := range podWithVolumeMountConflict.Spec.Containers {
		podWithVolumeMountConflict.Spec.Containers[i].VolumeMounts = append(container.VolumeMounts, volumeMounts...)
	}
	podWithVolumeMountConflictRaw, _ := json.Marshal(podWithVolumeMountConflict)
	admissionReviewWithVolumeMountConflict.Request.Object.Raw = podWithVolumeMountConflictRaw
	admissionReviewWithVolumeMountConflictRequest, _ := json.Marshal(admissionReviewWithVolumeMountConflict)

	admissionReviewWithVolumeConflict := admissionv1.AdmissionReview{}
	admissionReviewExample.DeepCopyInto(&admissionReviewWithVolumeConflict)
	var podWithVolumeConflict corev1.Pod
	if err := json.Unmarshal(admissionReviewWithVolumeConflict.Request.Object.Raw, &podWithVolumeConflict); err != nil {
		t.Error(err)
	}
	volume := volumesTemplate[0:1]
	podWithVolumeConflict.Spec.Volumes = append(podWithVolumeConflict.Spec.Volumes, volume...)
	podWithVolumeConflictRaw, _ := json.Marshal(podWithVolumeConflict)
	admissionReviewWithVolumeConflict.Request.Object.Raw = podWithVolumeConflictRaw
	admissionReviewWithVolumeConflictRequest, _ := json.Marshal(admissionReviewWithVolumeConflict)

	defaultReqHeader := http.Header{"Content-Type": {"application/json"}}
	testCases := []struct {
		name     string
		method   string
		header   map[string][]string
		reqBody  string
		httpCode int
		respBody string
	}{
		{"test with example data", http.MethodPost, defaultReqHeader, string(admissionReviewExampleRequest), http.StatusOK, exceptMutatedJsonPatch},
		{"test with kube-system namespace", http.MethodPost, defaultReqHeader, string(admissionReviewWithNamespaceSystemRequest), http.StatusOK, exceptSkipJsonPatch},
		{"test with scale kind", http.MethodPost, defaultReqHeader, string(admissionReviewWithScaleKindRequest), http.StatusOK, exceptSkipJsonPatch},
		{"test with deny annotation", http.MethodPost, defaultReqHeader, string(admissionReviewWithDenyAnnotationRequest), http.StatusOK, exceptReplaceSkipJsonPatch},
		{"test with volume mount conflict", http.MethodPost, defaultReqHeader, string(admissionReviewWithVolumeMountConflictRequest), http.StatusOK, exceptConflictJsonPatch},
		{"test with volume conflict", http.MethodPost, defaultReqHeader, string(admissionReviewWithVolumeConflictRequest), http.StatusOK, exceptConflictJsonPatch},
	}

	for _, testCase := range testCases {
		t.Logf("Test case for: %s", testCase.name)
		req, err := http.NewRequest(testCase.method, "/mutate", strings.NewReader(testCase.reqBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header = testCase.header

		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(whsvr.serve)

		handler.ServeHTTP(recorder, req)

		responseAdmissionReview := admissionv1.AdmissionReview{}
		responseJsonPatch := ""
		if _, _, err := deserializer.Decode(recorder.Body.Bytes(), nil, &responseAdmissionReview); err != nil {
			t.Errorf("Can't decode response body: %v", err)
		}

		patchTypeJSONPatch := admissionv1.PatchTypeJSONPatch
		if *responseAdmissionReview.Response.PatchType == patchTypeJSONPatch {
			responseJsonPatch = string(responseAdmissionReview.Response.Patch)
		}

		assert.Equal(t, recorder.Code, testCase.httpCode, fmt.Sprintf("handler returned wrong status code: got %v want %v", recorder.Code, testCase.httpCode))
		assert.Equal(t, strings.Contains(responseJsonPatch, testCase.respBody), true, fmt.Sprintf("handler returned unexpected body: got %v not contain %v", responseJsonPatch, testCase.respBody))
	}
}

func TestStartWebhookServer(t *testing.T) {
	parameters := WhSvrParameters{
		8443,
		"../deploy/certs/server-cert.pem",
		"../deploy/certs/server-key.pem",
	}

	whsvr := startWebhookServer(&parameters)
	defer func() {
		_ = whsvr.server.Close()
	}()

	exceptRespBody := "pong"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get("https://localhost:8443/ping")
	if err != nil {
		t.Error(err)
	}
	respBodyByte, err := ioutil.ReadAll(resp.Body)
	respBody := string(respBodyByte)
	if respBody != exceptRespBody {
		t.Errorf("got unexpected body: got %v ,except %v", respBody, exceptRespBody)
	}
}

func TestMutationRequired(t *testing.T) {
	admissionReviewExample := GetAdmissionReviewExample()

	NotRequiredCase1 := admissionReviewExample.DeepCopy()
	NotRequiredCase1.Request.Namespace = metav1.NamespaceSystem

	NotRequiredCase2 := admissionReviewExample.DeepCopy()
	NotRequiredCase2.Request.Namespace = metav1.NamespacePublic

	NotRequiredCase3 := admissionReviewExample.DeepCopy()
	var podWithDenyAnnotation corev1.Pod
	if err := json.Unmarshal(NotRequiredCase3.Request.Object.Raw, &podWithDenyAnnotation); err != nil {
		t.Error(err)
	}
	podWithDenyAnnotation.SetAnnotations(map[string]string{
		admissionWebhookAnnotationEnableKey: "No",
		admissionWebhookAnnotationStatusKey: "test",
	})
	podWithDenyAnnotationRaw, _ := json.Marshal(podWithDenyAnnotation)
	NotRequiredCase3.Request.Object.Raw = podWithDenyAnnotationRaw

	NotRequiredCase4 := admissionReviewExample.DeepCopy()
	var podWithMutatedAnnotation corev1.Pod
	if err := json.Unmarshal(NotRequiredCase4.Request.Object.Raw, &podWithMutatedAnnotation); err != nil {
		t.Error(err)
	}
	podWithMutatedAnnotation.SetAnnotations(map[string]string{
		admissionWebhookAnnotationStatusKey: admissionWebhookSuccessFlag,
	})
	podWithMutatedAnnotationRaw, _ := json.Marshal(podWithMutatedAnnotation)
	NotRequiredCase4.Request.Object.Raw = podWithMutatedAnnotationRaw

	RequiredCase5 := admissionReviewExample.DeepCopy()
	RequiredCase5.Request.Kind = validMutatingKindList[0]

	NotRequiredCase5 := admissionReviewExample.DeepCopy()
	NotRequiredCase5.Request.Kind = metav1.GroupVersionKind{
		Group:   "autoscaling",
		Version: "v1",
		Kind:    "Scale",
	}

	RequiredCase6 := admissionReviewExample.DeepCopy()
	RequiredCase6.Request.Operation = validMutatingOperationList[0]

	NotRequiredCase6 := admissionReviewExample.DeepCopy()
	NotRequiredCase6.Request.Operation = admissionv1.Update

	cases := []struct {
		admissionReview *admissionv1.AdmissionReview
		required        bool
	}{
		{NotRequiredCase1, false},
		{NotRequiredCase2, false},
		{NotRequiredCase3, false},
		{NotRequiredCase4, false},
		{NotRequiredCase5, false},
		{RequiredCase5, true},
		{NotRequiredCase6, false},
		{RequiredCase6, true},
	}

	for _, testCase := range cases {
		assert.Equal(t, mutationRequired(ignoredNamespaces, validMutatingKindList, validMutatingOperationList, testCase.admissionReview), testCase.required)
	}
}

func TestEscapeJSONPointerValue(t *testing.T) {
	origin := "{\"foo/bar~\": \"baz\"}"
	except := "{\"foo~1bar~0\": \"baz\"}"
	escape := escapeJSONPointerValue(origin)
	assert.Equal(t, escape, except, fmt.Sprintf("TestEscapeJSONPointerValue: got %v want %v", escape, except))
}
