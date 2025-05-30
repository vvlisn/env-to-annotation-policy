package main

import (
	"encoding/json"
	"testing"

	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

func TestValidSettings(t *testing.T) {
	settings := Settings{
		EnvKey:              "test_env",
		AnnotationBase:      "test_base",
		AnnotationExtFormat: "test_ext_%d",
	}

	valid, err := settings.Valid()
	if !valid {
		t.Errorf("Expected settings to be valid, got error: %v", err)
	}
}

func TestInvalidSettingsEmptyEnvKey(t *testing.T) {
	settings := Settings{
		EnvKey:              "",
		AnnotationBase:      "test_base",
		AnnotationExtFormat: "test_ext_%d",
	}

	valid, err := settings.Valid()
	if valid {
		t.Errorf("Expected settings to be invalid due to empty EnvKey")
	}
	if err == nil || err.Error() != "env_key cannot be empty" {
		t.Errorf("Expected error 'env_key cannot be empty', got: %v", err)
	}
}

func TestInvalidSettingsEmptyAnnotationBase(t *testing.T) {
	settings := Settings{
		EnvKey:              "test_env",
		AnnotationBase:      "",
		AnnotationExtFormat: "test_ext_%d",
	}

	valid, err := settings.Valid()
	if valid {
		t.Errorf("Expected settings to be invalid due to empty AnnotationBase")
	}
	if err == nil || err.Error() != "annotation_base cannot be empty" {
		t.Errorf("Expected error 'annotation_base cannot be empty', got: %v", err)
	}
}

func TestInvalidSettingsEmptyAnnotationExtFormat(t *testing.T) {
	settings := Settings{
		EnvKey:              "test_env",
		AnnotationBase:      "test_base",
		AnnotationExtFormat: "",
	}

	valid, err := settings.Valid()
	if valid {
		t.Errorf("Expected settings to be invalid due to empty AnnotationExtFormat")
	}
	if err == nil || err.Error() != "annotation_ext_format cannot be empty" {
		t.Errorf("Expected error 'annotation_ext_format cannot be empty', got: %v", err)
	}
}

func TestInvalidSettingsAnnotationExtFormatMissingPlaceholder(t *testing.T) {
	settings := Settings{
		EnvKey:              "test_env",
		AnnotationBase:      "test_base",
		AnnotationExtFormat: "test_ext",
	}

	valid, err := settings.Valid()
	if valid {
		t.Errorf("Expected settings to be invalid due to missing %%d placeholder in AnnotationExtFormat")
	}
	if err == nil || err.Error() != "annotation_ext_format must contain %d placeholder" {
		t.Errorf("Expected error 'annotation_ext_format must contain %%d placeholder', got: %v", err)
	}
}

func TestNewSettingsFromValidationReqWithValidSettings(t *testing.T) {
	rawSettings := []byte(`{"env_key": "my_env", "annotation_base": "my_base", "annotation_ext_format": "my_ext_%d"}`)
	validationReq := &kubewarden_protocol.ValidationRequest{
		Settings: rawSettings,
	}

	settings, err := NewSettingsFromValidationReq(validationReq)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if settings.EnvKey != "my_env" {
		t.Errorf("Expected EnvKey to be 'my_env', got '%s'", settings.EnvKey)
	}
	if settings.AnnotationBase != "my_base" {
		t.Errorf("Expected AnnotationBase to be 'my_base', got '%s'", settings.AnnotationBase)
	}
	if settings.AnnotationExtFormat != "my_ext_%d" {
		t.Errorf("Expected AnnotationExtFormat to be 'my_ext_%%d', got '%s'", settings.AnnotationExtFormat)
	}
}

func TestNewSettingsFromValidationReqWithInvalidJSON(t *testing.T) {
	rawSettings := []byte(`{"env_key": "my_env", "annotation_base": "my_base", "annotation_ext_format": "my_ext_%d"`)
	validationReq := &kubewarden_protocol.ValidationRequest{
		Settings: rawSettings,
	}

	_, err := NewSettingsFromValidationReq(validationReq)
	if err == nil {
		t.Errorf("Expected an error due to invalid JSON, got nil")
	}
}

func TestSettingsUnmarshalWithNoValueProvided(t *testing.T) {
	rawSettings := []byte(`{}`)
	settings := &Settings{}
	if err := json.Unmarshal(rawSettings, settings); err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if settings.EnvKey != "" {
		t.Errorf("Expected EnvKey to be empty, got '%s'", settings.EnvKey)
	}
	if settings.AnnotationBase != "" {
		t.Errorf("Expected AnnotationBase to be empty, got '%s'", settings.AnnotationBase)
	}
	if settings.AnnotationExtFormat != "" {
		t.Errorf("Expected AnnotationExtFormat to be empty, got '%s'", settings.AnnotationExtFormat)
	}

	valid, err := settings.Valid()
	if valid {
		t.Errorf("Expected settings to be invalid, got valid")
	}
	if err == nil {
		t.Errorf("Expected an error when settings are invalid, got nil")
	}
}
