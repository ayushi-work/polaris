package iforge

// Incident types.
const (
	IncidentTypeOOMKilled        = "OOMKilled"
	IncidentTypeCrashLoopBackOff = "CrashLoopBackOff"
	IncidentTypeImagePullBackOff = "ImagePullBackOff"
	IncidentTypeErrImagePull     = "ErrImagePull"
	IncidentTypePodPending       = "PodPending"
	IncidentTypeNodePressure     = "NodePressure"
	IncidentTypeDeployRollout    = "DeployRolloutStuck"
	IncidentTypeHighRestartCount = "HighRestartCount"
)

// Incident statuses.
const (
	IncidentStatusDetected      = "detected"
	IncidentStatusInvestigating = "investigating"
	IncidentStatusRemediating   = "remediating"
	IncidentStatusResolved      = "resolved"
	IncidentStatusFailed        = "failed"
)

// Severity levels.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Remediation types.
const (
	RemediationTypeRestart   = "restart"
	RemediationTypeScaleUp   = "scale_up"
	RemediationTypeScaleDown = "scale_down"
	RemediationTypeRollback  = "rollback"
	RemediationTypeDrainNode = "drain_node"
	RemediationTypeCordonNode = "cordon_node"
	RemediationTypeDeletePod = "delete_pod"
)

// Remediation statuses.
const (
	RemediationPending  = "pending"
	RemediationRunning  = "running"
	RemediationSuccess  = "success"
	RemediationFailed   = "failed"
	RemediationSkipped  = "skipped"
)

// Chaos actions.
const (
	ChaosActionDeletePod    = "delete_pod"
	ChaosActionStressCPU    = "stress_cpu"
	ChaosActionStressMemory = "stress_memory"
	ChaosActionNetworkDelay = "network_delay"
	ChaosActionNetworkLoss  = "network_loss"
	ChaosActionConfigFault  = "config_fault"
)
