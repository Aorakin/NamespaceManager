package httpclient

import (
	"fmt"
	"net/http"

	apiError "github.com/NamespaceManager/pkg/api_error"
)

// handleResponse maps HTTP status codes to appropriate ApiErrors
// Returns response body for 2xx status codes, otherwise returns mapped error
func handleResponse(statusCode int, body []byte) ([]byte, error) {
	// Success cases (2xx)
	if statusCode >= 200 && statusCode < 300 {
		return body, nil
	}

	// Map specific status codes to API errors
	switch statusCode {
	case http.StatusBadRequest:
		return nil, apiError.NewBadRequestError(string(body))
	case http.StatusUnauthorized:
		return nil, apiError.NewUnauthorizedError(string(body))
	case http.StatusForbidden:
		return nil, apiError.NewForbiddenError(string(body))
	case http.StatusNotFound:
		return nil, apiError.NewNotFoundError(string(body))
	case http.StatusConflict:
		return nil, apiError.NewConflictError(string(body))
	case http.StatusTooManyRequests:
		return nil, apiError.NewTooManyRequestsError(string(body))
	case http.StatusServiceUnavailable:
		return nil, apiError.NewServiceUnavailableError(string(body))
	default:
		// Handle 5xx server errors
		if statusCode >= 500 {
			return nil, apiError.NewInternalServerError(fmt.Sprintf("external service error: %s", string(body)))
		}
		// Handle other client errors (4xx)
		return nil, apiError.NewApiError(statusCode, "external service error", string(body))
	}
}
