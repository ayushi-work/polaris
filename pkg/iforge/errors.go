package iforge

// APIError represents a structured API error response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

var (
	ErrIncidentNotFound    = &APIError{Code: "INCIDENT_NOT_FOUND", Message: "Incident not found"}
	ErrRemediationNotFound = &APIError{Code: "REMEDIATION_NOT_FOUND", Message: "Remediation not found"}
	ErrInvalidInput        = &APIError{Code: "INVALID_INPUT", Message: "Invalid request parameters"}
	ErrClusterUnreachable  = &APIError{Code: "CLUSTER_UNREACHABLE", Message: "Cannot connect to Kubernetes cluster"}
	ErrLLMUnavailable      = &APIError{Code: "LLM_UNAVAILABLE", Message: "LLM provider unreachable"}
	ErrInternal            = &APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrScenarioNotFound    = &APIError{Code: "SCENARIO_NOT_FOUND", Message: "Chaos scenario not found"}
	ErrRCAAlreadyExists    = &APIError{Code: "RCA_ALREADY_EXISTS", Message: "RCA result already exists for this incident"}
)
