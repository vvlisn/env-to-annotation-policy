package main

import (
	"encoding/json"
	"fmt"

	appsv1 "github.com/kubewarden/k8s-objects/api/apps/v1"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
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

	// 提取 env 中目标 key 的值.
	logPaths := extractLogPathsFromContainers(deployment.Spec.Template.Spec.Containers, settings.EnvKey)
	if len(logPaths) == 0 {
		return kubewarden.AcceptRequest()
	}

	// 构造注解 map.
	annotations := buildAnnotations(logPaths, settings)

	// 将注解设置到 metadata.annotations 中.
	if deployment.Metadata.Annotations == nil {
		deployment.Metadata.Annotations = map[string]string{}
	}
	for k, v := range annotations {
		deployment.Metadata.Annotations[k] = v
	}

	return kubewarden.MutateRequest(deployment)
}

// extractLogPathsFromContainers 提取所有容器中目标 env_key 的值.
func extractLogPathsFromContainers(containers []*corev1.Container, targetKey string) []string {
	var results []string
	for _, container := range containers {
		for _, env := range container.Env {
			if *env.Name == targetKey {
				results = append(results, env.Value)
			}
		}
	}
	return results
}

// buildAnnotations 构造注解键值对.
func buildAnnotations(logPaths []string, settings Settings) map[string]string {
	annotations := make(map[string]string)
	for i, path := range logPaths {
		if i == 0 {
			annotations[settings.AnnotationBase] = path
		} else {
			key := fmt.Sprintf(settings.AnnotationExtFormat, i)
			annotations[key] = path
		}
	}
	return annotations
}
