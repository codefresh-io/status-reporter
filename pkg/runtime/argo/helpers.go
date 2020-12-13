package argo

import (
	"fmt"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
)

func hasWorkflowStarted(wf *wfv1.Workflow) bool {
	switch wf.Status.Phase {
	case "", wfv1.NodePending:
		return false
	default:
		return true
	}
}

func hasWorkflowFinished(wf *wfv1.Workflow) bool {
	switch wf.Status.Phase {
	case wfv1.NodeFailed, wfv1.NodeSucceeded:
		return true
	default:
		return false
	}
}

func hasWorkflowFailed(wf *wfv1.Workflow) error {
	if wf.Status.Phase == wfv1.NodeFailed {
		return fmt.Errorf("Workflow %v has failed: %s", wf.ObjectMeta.Name, wf.Status.Message)
	}
	return nil
}

func hasStepStatusChanged(s *reporter.WorkflowStep, n *wfv1.NodeStatus) bool {
	return getStepStatus(n) != s.Status
}

func getStepStatus(n *wfv1.NodeStatus) reporter.WorkflowStepStatus {
	if !hasStepStarted(n) {
		return reporter.WorkflowStepPending
	}

	if !hasStepFinished(n) {
		return reporter.WorkflowStepRunning
	}

	if wasStepSkipped(n) {
		return reporter.WorkflowStepSkipped
	}

	if wasStepSuccessful(n) {
		return reporter.WorkflowStepSucceded
	}

	return reporter.WorkflowStepFailed
}

func hasStepStarted(n *wfv1.NodeStatus) bool {
	return !n.StartedAt.IsZero()
}

func hasStepFinished(n *wfv1.NodeStatus) bool {
	return n.Completed()
}

func wasStepSkipped(n *wfv1.NodeStatus) bool {
	return n.Phase == wfv1.NodeSkipped
}

func wasStepSuccessful(n *wfv1.NodeStatus) bool {
	return n.Successful()
}

func hasStepFailed(n *wfv1.NodeStatus) bool {
	return !n.Successful()
}

func getStepError(n *wfv1.NodeStatus) error {
	return fmt.Errorf("step %s failed with: %s", n.DisplayName, n.Message)
}
