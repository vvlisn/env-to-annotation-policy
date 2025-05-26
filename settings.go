package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

// Settings 定义了策略中的所有可配置项.
type Settings struct {
	// EnvKey 容器环境变量名称，用于匹配需要转换的环境变量
	EnvKey string `json:"env_key"`
	// AnnotationBase 基础注解键名，用于第一个日志路径
	AnnotationBase string `json:"annotation_base"`
	// AnnotationExtFormat 扩展注解键名格式，用于后续的日志路径
	// 格式为: co_elastic_logs_path_ext_%d，其中 %d 会被替换为序号 1,2,3...
	AnnotationExtFormat string `json:"annotation_ext_format"`
}

// NewSettingsFromValidationReq 从 ValidationRequest 中提取设置.
func NewSettingsFromValidationReq(validationReq *kubewarden_protocol.ValidationRequest) (Settings, error) {
	settings := Settings{}
	err := json.Unmarshal(validationReq.Settings, &settings)
	return settings, err
}

// Valid 对 Settings 本身做合法性校验.
func (s *Settings) Valid() (bool, error) {
	if s.EnvKey == "" {
		return false, errors.New("env_key cannot be empty")
	}
	if s.AnnotationBase == "" {
		return false, errors.New("annotation_base cannot be empty")
	}
	if s.AnnotationExtFormat == "" {
		return false, errors.New("annotation_ext_format cannot be empty")
	}

	// 验证 AnnotationExtFormat 是否包含格式化占位符 %d
	if !strings.Contains(s.AnnotationExtFormat, "%d") {
		return false, errors.New("annotation_ext_format must contain %d placeholder")
	}
	return true, nil
}

// validateSettings 由 Kubewarden 在策略加载时调用.
func validateSettings(payload []byte) ([]byte, error) {
	logger.Info("validating settings")

	settings := Settings{}
	err := json.Unmarshal(payload, &settings)
	if err != nil {
		return kubewarden.RejectSettings(kubewarden.Message(fmt.Sprintf("Provided settings are not valid: %v", err)))
	}

	valid, err := settings.Valid()
	if err != nil {
		return kubewarden.RejectSettings(kubewarden.Message(fmt.Sprintf("Provided settings are not valid: %v", err)))
	}
	if valid {
		return kubewarden.AcceptSettings()
	}

	logger.Warn("rejecting settings")
	return kubewarden.RejectSettings(kubewarden.Message("Provided settings are not valid"))
}
