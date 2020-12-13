package tekton

import (
	"fmt"

	"github.com/codefresh-io/status-reporter/pkg/reporter"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

func PipelineHasStarted(pr *v1beta1.PipelineRun) bool {
	if pr.Status.Conditions == nil {
		return false
	}
	return len(pr.Status.Conditions) != 0
}

func PipelineHasFinished(pr *v1beta1.PipelineRun) bool {
	if len(pr.Status.Conditions) == 0 {
		return false
	}

	return pr.Status.Conditions[0].Status != corev1.ConditionUnknown
}

func PipelineHasFailed(pr *v1beta1.PipelineRun) error {
	prConditions := pr.Status.Conditions
	if len(prConditions) != 0 && prConditions[0].Status == corev1.ConditionFalse {
		return fmt.Errorf("pipelinerun has failed: %s", prConditions[0].Message)
	}
	return nil
}

func PipelineIsRunning(pr *v1beta1.PipelineRun) bool {
	prConditions := pr.Status.Conditions
	return len(prConditions) != 0 && prConditions[0].Status == corev1.ConditionUnknown
}

func GetPipelineState(pr *v1beta1.PipelineRun) (reporter.WorkflowStatus, error) {
	if !PipelineHasStarted(pr) {
		return reporter.WorkflowPending, nil
	}

	if PipelineIsRunning(pr) {
		return reporter.WorkflowRunning, nil
	}

	if PipelineHasFinished(pr) {
		if PipelineHasFailed(pr) != nil {
			return reporter.WorkflowFailed, nil
		}
		return reporter.WorkflowSucceded, nil
	}
	return reporter.WorkflowStatus(""), fmt.Errorf("unknown pipeline state")
}

func TaskHasStarted(trs *v1alpha1.PipelineRunTaskRunStatus) bool {
	return trs.Status.StartTime != nil && !trs.Status.StartTime.IsZero()
}

func TaskHasFinished(trs *v1alpha1.PipelineRunTaskRunStatus) bool {
	return !trs.Status.GetCondition(apis.ConditionSucceeded).IsUnknown()
}

func TaskHasFailed(trs *v1alpha1.PipelineRunTaskRunStatus) error {
	trsConditions := trs.Status.Conditions
	if len(trsConditions) != 0 && trsConditions[0].Status == corev1.ConditionFalse {
		return fmt.Errorf("taskrun has failed: %s", trsConditions[0].Message)
	}
	return nil
}

func TaskIsSuccessful(trs *v1alpha1.PipelineRunTaskRunStatus) bool {
	return trs.Status.GetCondition(apis.ConditionSucceeded).IsTrue()
}

func GetTaskStatus(trs *v1alpha1.PipelineRunTaskRunStatus) (reporter.WorkflowStepStatus, error) {
	if !TaskHasStarted(trs) {
		return reporter.WorkflowStepPending, nil
	}

	if !TaskHasFinished(trs) {
		return reporter.WorkflowStepRunning, nil
	}

	if TaskIsSuccessful(trs) {
		return reporter.WorkflowStepSucceded, nil
	}

	if TaskHasFailed(trs) != nil {
		return reporter.WorkflowStepFailed, nil
	}

	return reporter.WorkflowStepStatus(""), fmt.Errorf("unknown task state")
}

func HasStepStatusChanged(step *reporter.WorkflowStep, trs *v1alpha1.PipelineRunTaskRunStatus) bool {
	if newState, err := GetTaskStatus(trs); err != nil {
		return false
	} else {
		return step.Status != newState
	}
}
