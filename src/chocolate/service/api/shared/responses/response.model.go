package responses

import (
	"chocolate/service/api/shared/apierror"
)

// APIResponse wrapper on top of every request
type APIResponse struct {
	Data  interface{}        `json:"data,omitempty"`
	Error *apierror.Error `json:"error,omitempty"`
}
