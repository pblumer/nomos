package app

import (
	"errors"
	"fmt"
)

const (
	CodeCosmosNotFound         = "COSMOS_NOT_FOUND"
	CodeCosmosMissing          = "COSMOS_MISSING"
	CodeCosmosLoadFailed       = "COSMOS_LOAD_FAILED"
	CodeDomainNotFound         = "DOMAIN_NOT_FOUND"
	CodeServiceNotFound        = "SERVICE_NOT_FOUND"
	CodeBlueprintNotFound      = "BLUEPRINT_NOT_FOUND"
	CodeInstanceNotFound       = "INSTANCE_NOT_FOUND"
	CodeValidationFailed       = "VALIDATION_FAILED"
	CodeInvalidFormat          = "INVALID_FORMAT"
	CodeInvalidNamespace       = "INVALID_NAMESPACE"
	CodeInternalError          = "INTERNAL_ERROR"
	CodeInvalidInput           = "INVALID_INPUT"
	CodeBlueprintAlreadyExists = "BLUEPRINT_ALREADY_EXISTS"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Cause      error  `json:"-"`
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }
func (e *AppError) Unwrap() error { return e.Cause }

func Error(code, message string, status int, cause error) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: status, Cause: cause}
}

func ErrorResponse(err error) map[string]any {
	if ae, ok := AsAppError(err); ok {
		return map[string]any{"error": map[string]string{"code": ae.Code, "message": ae.Message}}
	}
	return map[string]any{"error": map[string]string{"code": CodeInternalError, "message": err.Error()}}
}

func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
