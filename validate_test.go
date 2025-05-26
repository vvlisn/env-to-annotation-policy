package main

import (
	"encoding/json"
	"fmt"
	"testing"

	appsv1 "github.com/kubewarden/k8s-objects/api/apps/v1"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	metav1 "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

// validateTest 是一个辅助函数，用于执行策略验证.
func validateTest(
	t *testing.T,
	request kubewarden_protocol.ValidationRequest,
) (*kubewarden_protocol.ValidationResponse, error) {
	t.Helper()

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	responsePayload, err := validate(payload)
	if err != nil {
		return nil, err
	}

	var response kubewarden_protocol.ValidationResponse
	if unmarshalErr := json.Unmarshal(responsePayload, &response); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &response, nil
}

func TestDeploymentMutation(t *testing.T) {
	tests := []struct {
		name                string
		settings            Settings
		deployment          appsv1.Deployment
		expectedAnnotations map[string]string
		shouldMutate        bool
	}{
		{
			name: "deployment with single container and target env",
			settings: Settings{
				EnvKey:              "vestack_varlog",
				AnnotationBase:      "co_elastic_logs_path",
				AnnotationExtFormat: "co_elastic_logs_path_ext_%d",
			},
			deployment: appsv1.Deployment{
				Spec: &appsv1.DeploymentSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: &corev1.PodSpec{
							Containers: []*corev1.Container{
								{
									Name: stringPtr("my-container"), // 添加容器名称
									Env: []*corev1.EnvVar{
										{Name: stringPtr("vestack_varlog"), Value: "/var/log/app.log"},
									},
								},
							},
						},
					},
				},
				Metadata: &metav1.ObjectMeta{}, // Initialize Metadata
			},
			expectedAnnotations: map[string]string{
				"my-container/co_elastic_logs_path": "/var/log/app.log",
			},
			shouldMutate: true,
		},
		{
			name: "deployment with multiple containers and target envs",
			settings: Settings{
				EnvKey:              "vestack_varlog",
				AnnotationBase:      "co_elastic_logs_path",
				AnnotationExtFormat: "co_elastic_logs_path_ext_%d",
			},
			deployment: appsv1.Deployment{
				Spec: &appsv1.DeploymentSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: &corev1.PodSpec{
							Containers: []*corev1.Container{
								{
									Name: stringPtr("container1"),
									Env: []*corev1.EnvVar{
										{Name: stringPtr("vestack_varlog"), Value: "/var/log/app1.log"},
									},
								},
								{
									Name: stringPtr("container2"),
									Env: []*corev1.EnvVar{
										{Name: stringPtr("vestack_varlog"), Value: "/var/log/app2.log"},
									},
								},
							},
						},
					},
				},
				Metadata: &metav1.ObjectMeta{}, // Initialize Metadata
			},
			expectedAnnotations: map[string]string{
				"container1/co_elastic_logs_path": "/var/log/app1.log",
				"container2/co_elastic_logs_path": "/var/log/app2.log",
			},
			shouldMutate: true,
		},
		{
			name: "deployment with no target env",
			settings: Settings{
				EnvKey:              "vestack_varlog",
				AnnotationBase:      "co_elastic_logs_path",
				AnnotationExtFormat: "co_elastic_logs_path_ext_%d",
			},
			deployment: appsv1.Deployment{
				Spec: &appsv1.DeploymentSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: &corev1.PodSpec{
							Containers: []*corev1.Container{
								{
									Name: stringPtr("my-container"),
									Env: []*corev1.EnvVar{
										{Name: stringPtr("OTHER_ENV"), Value: "some_value"},
									},
								},
							},
						},
					},
				},
				Metadata: &metav1.ObjectMeta{}, // Initialize Metadata
			},
			expectedAnnotations: map[string]string{}, // No mutation expected, so empty map
			shouldMutate:        false,
		},
		{
			name: "deployment with existing annotations",
			settings: Settings{
				EnvKey:              "vestack_varlog",
				AnnotationBase:      "co_elastic_logs_path",
				AnnotationExtFormat: "co_elastic_logs_path_ext_%d",
			},
			deployment: appsv1.Deployment{
				Metadata: &metav1.ObjectMeta{
					Annotations: map[string]string{
						"existing_annotation": "value",
					},
				},
				Spec: &appsv1.DeploymentSpec{
					Template: &corev1.PodTemplateSpec{
						Metadata: &metav1.ObjectMeta{ // Initialize PodTemplateSpec Metadata
							Annotations: map[string]string{
								"existing_template_annotation": "template_value",
							},
						},
						Spec: &corev1.PodSpec{
							Containers: []*corev1.Container{
								{
									Name: stringPtr("my-container"),
									Env: []*corev1.EnvVar{
										{Name: stringPtr("vestack_varlog"), Value: "/var/log/test.log"},
									},
								},
							},
						},
					},
				},
			},
			expectedAnnotations: map[string]string{
				"existing_template_annotation":      "template_value",
				"my-container/co_elastic_logs_path": "/var/log/test.log",
			},
			shouldMutate: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runTest(t, test)
		})
	}
}

func runTest(t *testing.T, test struct {
	name                string
	settings            Settings
	deployment          appsv1.Deployment
	expectedAnnotations map[string]string
	shouldMutate        bool
}) {
	t.Helper()
	req := kubewarden_protocol.ValidationRequest{
		Request: kubewarden_protocol.KubernetesAdmissionRequest{
			Kind: kubewarden_protocol.GroupVersionKind{
				Kind: "Deployment",
			},
			Object: json.RawMessage(mustMarshalJSON(test.deployment)),
		},
		Settings: json.RawMessage(mustMarshalJSON(test.settings)),
	}

	response, err := validateTest(t, req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if test.shouldMutate {
		assertMutation(t, response, test.expectedAnnotations)
	} else {
		assertNoMutation(t, response)
	}
}

func assertMutation(
	t *testing.T,
	response *kubewarden_protocol.ValidationResponse,
	expectedAnnotations map[string]string,
) {
	t.Helper()
	if response.MutatedObject == nil {
		t.Errorf("Expected mutation, but MutatedObject is nil. Message: %s", *response.Message)
	}
	var mutatedDeployment appsv1.Deployment
	mutatedObjectBytes, marshalErr := json.Marshal(response.MutatedObject)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal mutated object to bytes: %v", marshalErr)
	}
	if unmarshalErr := json.Unmarshal(mutatedObjectBytes, &mutatedDeployment); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal mutated object bytes: %v", unmarshalErr)
	}

	if mutatedDeployment.Spec.Template.Metadata == nil || mutatedDeployment.Spec.Template.Metadata.Annotations == nil {
		t.Errorf("Expected annotations to be present after mutation in Pod Template Metadata")
	} else {
		for k, v := range expectedAnnotations {
			if val, annotationOk := mutatedDeployment.Spec.Template.Metadata.Annotations[k]; !annotationOk || val != v {
				t.Errorf("Expected annotation %s=%s, got %s=%s", k, v, k, val)
			}
		}
		if len(mutatedDeployment.Spec.Template.Metadata.Annotations) != len(expectedAnnotations) {
			t.Errorf(
				"Expected %d annotations, got %d",
				len(expectedAnnotations),
				len(mutatedDeployment.Spec.Template.Metadata.Annotations),
			)
		}
	}
}

func assertNoMutation(t *testing.T, response *kubewarden_protocol.ValidationResponse) {
	t.Helper()
	if response.MutatedObject != nil {
		t.Errorf("Expected no mutation, but got mutation")
	}
	if !response.Accepted {
		t.Errorf("Expected request to be accepted, but got rejected. Message: %s", *response.Message)
	}
}

// stringPtr 返回一个指向给定字符串的指针.
func stringPtr(s string) *string {
	return &s
}

// mustMarshalJSON 是一个辅助函数，用于将 Go 对象序列化为 JSON 字节数组.
func mustMarshalJSON(obj interface{}) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return data
}
