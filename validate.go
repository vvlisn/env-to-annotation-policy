package main

import (
	"encoding/json"
	"fmt"

	appsv1 "github.com/kubewarden/k8s-objects/api/apps/v1"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1" // 导入 corev1
	metav1 "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

const RejectCode = 400

// validate 是入口函数.
func validate(payload []byte) ([]byte, error) {
	var validationRequest kubewarden_protocol.ValidationRequest
	if err := json.Unmarshal(payload, &validationRequest); err != nil {
		return kubewarden.RejectRequest(kubewarden.Message(err.Error()), kubewarden.Code(RejectCode))
	}

	settings, err := NewSettingsFromValidationReq(&validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(kubewarden.Message(err.Error()), kubewarden.Code(RejectCode))
	}

	return processDeployment(validationRequest, settings)
}

// processDeployment 处理 Deployment 类型的资源.
func processDeployment(req kubewarden_protocol.ValidationRequest, settings Settings) ([]byte, error) {
	if req.Request.Kind.Kind != "Deployment" {
		return kubewarden.AcceptRequest()
	}

	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.Request.Object, &deployment); err != nil {
		return kubewarden.RejectRequest(kubewarden.Message("cannot unmarshal deployment"), kubewarden.Code(RejectCode))
	}

	mutated := mutateDeploymentContainers(&deployment, settings)

	if !mutated {
		return kubewarden.AcceptRequest()
	}

	return kubewarden.MutateRequest(deployment)
}

func mutateDeploymentContainers(deployment *appsv1.Deployment, settings Settings) bool {
	mutated := false
	if deployment.Spec.Template.Metadata == nil {
		deployment.Spec.Template.Metadata = &metav1.ObjectMeta{}
	}
	if deployment.Spec.Template.Metadata.Annotations == nil {
		deployment.Spec.Template.Metadata.Annotations = map[string]string{}
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if processContainerEnv(container, deployment.Spec.Template.Metadata.Annotations, settings) {
			mutated = true
		}
	}
	return mutated
}

func processContainerEnv(container *corev1.Container, annotations map[string]string, settings Settings) bool {
	if container == nil {
		return false
	}
	var logPaths []string
	for _, env := range container.Env {
		if env == nil || env.Name == nil {
			continue
		}
		if *env.Name == settings.EnvKey {
			logPaths = append(logPaths, env.Value)
		}
	}

	if len(logPaths) > 0 {
		if container.Name == nil {
			// 如果容器没有名称，则无法为其添加带前缀的 annotation，跳过
			return false
		}
		for i, path := range logPaths {
			var annotationKey string
			if i == 0 {
				annotationKey = fmt.Sprintf("%s/%s", *container.Name, settings.AnnotationBase)
			} else {
				annotationKey = fmt.Sprintf("%s/%s", *container.Name, fmt.Sprintf(settings.AnnotationExtFormat, i))
			}
			annotations[annotationKey] = path
		}
		return true
	}
	return false
}
