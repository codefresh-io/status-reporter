package reporter

import "github.com/codefresh-io/status-reporter/pkg/logger"

type (
	// Reporter reports status of single step or pipeline
	Reporter interface {
		Report(string) error
	}

	// CodefreshAPI to report the status
	CodefreshAPI interface {
		ReportWorkflowStaus(workflow string, status string) error
		ReportWorkflowStepStaus(workflow string, step string, status string) error
	}

	// WorkflowStatusReporter implements Reporter
	WorkflowStatusReporter struct {
		CodefreshAPI CodefreshAPI
		Logger       logger.Logger
		WorkflowID   string
	}

	// WorkflowStepStatusReporter implements Reporter
	WorkflowStepStatusReporter struct {
		CodefreshAPI CodefreshAPI
		Logger       logger.Logger
		WorkflowID   string
		Step         string
	}
)

// Report status
func (w *WorkflowStatusReporter) Report(status string) error {
	w.Logger.Info("Reporting workflow status", "status", status, "workflow-id", w.WorkflowID)
	return w.CodefreshAPI.ReportWorkflowStaus(w.WorkflowID, status)
}

// Report status
func (w *WorkflowStepStatusReporter) Report(status string) error {
	w.Logger.Info("Reporting workflow status", "status", status, "workflow-id", w.WorkflowID, "step", w.Step)
	return w.CodefreshAPI.ReportWorkflowStaus(w.WorkflowID, status)
}
