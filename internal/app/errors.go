package app

import (
	"errors"
	"fmt"
)

const (
	CodeCosmosNotFound            = "COSMOS_NOT_FOUND"
	CodeCosmosMissing             = "COSMOS_MISSING"
	CodeCosmosLoadFailed          = "COSMOS_LOAD_FAILED"
	CodeDomainNotFound            = "DOMAIN_NOT_FOUND"
	CodeServiceNotFound           = "SERVICE_NOT_FOUND"
	CodeBlueprintNotFound         = "BLUEPRINT_NOT_FOUND"
	CodeInstanceNotFound          = "INSTANCE_NOT_FOUND"
	CodeServicegraphNotFound      = "SERVICEGRAPH_NOT_FOUND"
	CodeServicegraphAlreadyExists = "SERVICEGRAPH_ALREADY_EXISTS"
	CodeInstanceAlreadyExists     = "INSTANCE_ALREADY_EXISTS"
	CodeValidationFailed          = "VALIDATION_FAILED"
	CodeInvalidFormat             = "INVALID_FORMAT"
	CodeInvalidNamespace          = "INVALID_NAMESPACE"
	CodeInternalError             = "INTERNAL_ERROR"
	CodeInvalidInput              = "INVALID_INPUT"
	CodeBlueprintAlreadyExists    = "BLUEPRINT_ALREADY_EXISTS"
	CodeProductNotFound           = "PRODUCT_NOT_FOUND"
	CodeTargetDomainNotFound      = "TARGET_DOMAIN_NOT_FOUND"
	CodeProductMoveNoop           = "PRODUCT_MOVE_NOOP"
	CodeProductMoveInvalidTarget  = "PRODUCT_MOVE_INVALID_TARGET"
	CodeProductMoveWriteFailed    = "PRODUCT_MOVE_WRITE_FAILED"
	CodeRepositoryNotFound        = "REPOSITORY_NOT_FOUND"
	CodeMountNotFound             = "MOUNT_NOT_FOUND"
	CodeMountExists               = "MOUNT_EXISTS"
	CodeMountNotAuthenticated     = "MOUNT_NOT_AUTHENTICATED"
	CodeTypeNotFound              = "TYPE_NOT_FOUND"
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
