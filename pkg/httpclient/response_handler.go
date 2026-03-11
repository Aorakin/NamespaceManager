package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiError "github.com/NamespaceManager/pkg/api_error"
)

// extractErrorMessage tries to parse a JSON error body and return a human-readable message.
// It checks common fields: "message", "error", "detail". Falls back to the raw string.
func extractErrorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		// Not JSON — return raw string
		return string(body)
	}
	for _, key := range []string{"message", "error", "detail"} {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	// JSON but no recognised key — return raw string
	return string(body)
}

// handleResponse maps HTTP status codes to appropriate ApiErrors
// Returns response body for 2xx status codes, otherwise returns mapped error
func handleResponse(statusCode int, body []byte) ([]byte, error) {
	// Success cases (2xx)
	if statusCode >= 200 && statusCode < 300 {
		return body, nil
	}

	detail := extractErrorMessage(body)

	// Map specific status codes to API errors
	switch statusCode {
	case http.StatusBadRequest:
		return nil, apiError.NewBadRequestError(detail)
	case http.StatusUnauthorized:
		return nil, apiError.NewUnauthorizedError(detail)
	case http.StatusForbidden:
		return nil, apiError.NewForbiddenError(detail)
	case http.StatusNotFound:
		return nil, apiError.NewNotFoundError(detail)
	case http.StatusConflict:
		return nil, apiError.NewConflictError(detail)
	case http.StatusTooManyRequests:
		return nil, apiError.NewTooManyRequestsError(detail)
	case http.StatusServiceUnavailable:
		return nil, apiError.NewServiceUnavailableError(detail)
	default:
		// Handle 5xx server errors
		if statusCode >= 500 {
			return nil, apiError.NewInternalServerError(fmt.Sprintf("external service error: %s", detail))
		}
		// Handle other client errors (4xx)
		return nil, apiError.NewApiError(statusCode, "external service error", detail)
	}
}
