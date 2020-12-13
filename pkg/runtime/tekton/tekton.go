package tekton

import (
	"context"
	"fmt"

	"github.com/codefresh-io/status-reporter/pkg/codefresh"
	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
	"github.com/codefresh-io/status-reporter/pkg/runtime"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type re struct {
	logger    logger.Logger
	cs        *versioned.Clientset
	codefresh *codefresh.Codefresh
}

func New(opt *runtime.Options) (runtime.Runtime, error) {
	cs, err := versioned.NewForConfig(opt.Config)
	if err != nil {
		return nil, err
	}

	return &re{
		logger:    opt.Logger,
		cs:        cs,
		codefresh: opt.Client,
	}, nil
}

func (re *re) Watch(ctx context.Context, namespace, workflowID string) error {
	watch, err := re.cs.TektonV1beta1().PipelineRuns(namespace).Watch(ctx, v1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/pipeline=%s", workflowID),
		Watch:         true,
	})
	if err != nil {
		return err
	}
	defer watch.Stop()

	re.logger.Info("Watching tekton pipelines", "namespace", namespace)

	workflow := reporter.NewWorkflow()
	wsr := &reporter.WorkflowStatusReporter{
		CodefreshAPI: re.codefresh,
		Logger:       re.logger,
		WorkflowID:   workflowID,
	}

	for ev := range watch.ResultChan() {
		pr, ok := ev.Object.(*tkn.PipelineRun)
		if !ok {
			re.logger.Err(fmt.Errorf("Invalid object type"), "unexpected object type from event")
		}
		switch workflow.Status {
		case reporter.WorkflowPending:
			if handleWorkflowPending(workflow, pr, wsr) != nil {
				re.logger.Err(err, "failed to report workflow status")
			}
		case reporter.WorkflowRunning:
			if handleWorkflowRunning(workflow, pr, wsr) != nil {
				re.logger.Err(err, "failed to report workflow status")
			}
		}
		if PipelineHasFinished(pr) {
			if handleWorkflowFinished(workflow, pr, wsr) != nil {
				re.logger.Err(err, "failed to report workflow status")
			}
			break
		}
	}

	re.logger.Info("Workflow finished, exiting")
	return nil
}

func handleWorkflowPending(workflow *reporter.Workflow, pr *tkn.PipelineRun, wsr *reporter.WorkflowStatusReporter) error {
	if PipelineHasStarted(pr) {
		workflow.Status = reporter.WorkflowRunning
		for _, t := range pr.Status.TaskRuns {
			workflow.Steps[t.PipelineTaskName] = &reporter.WorkflowStep{Status: reporter.WorkflowStepPending}
		}
		return wsr.Report(reporter.WorkflowRunning, nil)
	}
	return nil
}

func handleWorkflowRunning(workflow *reporter.Workflow, pr *tkn.PipelineRun, wsr *reporter.WorkflowStatusReporter) error {
	for _, trs := range pr.Status.TaskRuns {
		if len(trs.Status.Steps) == 0 {
			wsr.Logger.Info("skipping task status report, steps are not running yet", "task", trs.PipelineTaskName)
			continue
		}
		stepName := trs.PipelineTaskName
		step := workflow.Steps[stepName]
		if step.Name == "" {
			step.Name = trs.Status.Steps[0].Name
		}
		if HasStepStatusChanged(step, trs) {
			newStatus, err := GetTaskStatus(trs)
			if err != nil {
				wsr.Logger.Err(err, "failed to get workflow step status")
				continue
			}
			step.Status = newStatus
			if newStatus == reporter.WorkflowStepFailed {
				if err = wsr.ReportStep(step.Name, reporter.WorkflowStepFailed, TaskHasFailed(trs)); err != nil {
					wsr.Logger.Err(err, "failed to report workflow step status")
				}
				continue
			}
			if err = wsr.ReportStep(step.Name, newStatus, nil); err != nil {
				wsr.Logger.Err(err, "failed to report workflow step status")
			}
		}
	}
	return nil
}

func handleWorkflowFinished(workflow *reporter.Workflow, pr *tkn.PipelineRun, wsr *reporter.WorkflowStatusReporter) error {
	if err := PipelineHasFailed(pr); err != nil {
		return wsr.Report(reporter.WorkflowFailed, err)
	}
	return wsr.Report(reporter.WorkflowSucceded, nil)
}
