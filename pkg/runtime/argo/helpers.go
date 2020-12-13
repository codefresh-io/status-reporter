package argo

import (
	"fmt"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
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
		if wf.Spec.Shutdown != "" {
			return fmt.Errorf("Workflow %v has been terminated", wf.ObjectMeta.Name)
		}
		return fmt.Errorf("Workflow %v has failed: %s", wf.ObjectMeta.Name, wf.Status.Message)
	}
	return nil
}
