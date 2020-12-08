package reporter

type (
	WorkflowStatus string

	WorkflowStepStatus string

	WorkflowStep struct {
		Name   string
		Status WorkflowStepStatus
	}

	Workflow struct {
		Status WorkflowStatus
		Steps  map[string]*WorkflowStep // maps task names to steps objects
	}
)

// Workflow statuses
const (
	WorkflowPending  WorkflowStatus = "pending"
	WorkflowRunning  WorkflowStatus = "running"
	WorkflowSucceded WorkflowStatus = "success"
	WorkflowFailed   WorkflowStatus = "error"
)

// Workflow step statuses
const (
	WorkflowStepPending  WorkflowStepStatus = "pending"
	WorkflowStepRunning  WorkflowStepStatus = "running"
	WorkflowStepSucceded WorkflowStepStatus = "success"
	WorkflowStepFailed   WorkflowStepStatus = "error"
	WorkflowStepSkipped  WorkflowStepStatus = "skipped"
)

func NewWorkflow() *Workflow {
	return &Workflow{
		Status: WorkflowPending,
		Steps:  map[string]*WorkflowStep{},
	}
}
