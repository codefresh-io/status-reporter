package reporter

import "github.com/codefresh-io/status-reporter/pkg/logger"

type (
	// Reporter reports status of single step or pipeline
	Reporter interface {
		Report(string) error
	}

	// CodefreshAPI to report the status
	CodefreshAPI interface {
		ReportWorkflowStaus(workflow string, status WorkflowStatus, err error) error
		ReportWorkflowStepStaus(workflow string, step string, status WorkflowStepStatus, err error) error
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
func (w *WorkflowStatusReporter) Report(status WorkflowStatus, err error) error {
	w.Logger.Info("Reporting workflow status", "workflow-id", w.WorkflowID, "status", status, "error", err)
	return w.CodefreshAPI.ReportWorkflowStaus(w.WorkflowID, status, err)
}

func (w *WorkflowStatusReporter) ReportStep(step string, status WorkflowStepStatus, err error) error {
	w.Logger.Info("Reporting workflow step status", "workflow-id", w.WorkflowID, "step", step, "status", status, "error", err)
	return w.CodefreshAPI.ReportWorkflowStepStaus(w.WorkflowID, step, status, err)
}

// Report status
func (w *WorkflowStepStatusReporter) Report(status WorkflowStepStatus) error {
	w.Logger.Info("Reporting workflow status", "status", status, "workflow-id", w.WorkflowID, "step", w.Step)
	return w.CodefreshAPI.ReportWorkflowStepStaus(w.WorkflowID, w.Step, status, nil)
}
