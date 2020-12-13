package argo

import (
	"context"
	"fmt"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	"github.com/argoproj/argo/workflow/packer"
	"github.com/codefresh-io/status-reporter/pkg/codefresh"
	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
	"github.com/codefresh-io/status-reporter/pkg/runtime"
	apierr "k8s.io/apimachinery/pkg/api/errors"
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
	wfIf := re.cs.ArgoprojV1alpha1().Workflows(namespace)
	opts := v1.ListOptions{
		LabelSelector: fmt.Sprintf("io.codefresh.processId=%s", workflowID),
		Watch:         true,
	}
	watch, err := wfIf.Watch(opts)
	if err != nil {
		return err
	}
	defer watch.Stop()

	re.logger.Info("Watching argo pipelines", "namespace", namespace)

	workflow := reporter.NewWorkflow()
	wsr := &reporter.WorkflowStatusReporter{
		CodefreshAPI: re.codefresh,
		Logger:       re.logger,
		WorkflowID:   workflowID,
	}

Loop:
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, open := <-watch.ResultChan():
			if !open {
				re.logger.Info("Re-establishing workflow watch")
				watch.Stop()
				watch, err = wfIf.Watch(opts)
				if err != nil {
					return err
				}
				continue
			}
			re.logger.Info("Received workflow event")
			wf, ok := event.Object.(*wfv1.Workflow)
			if !ok {
				// object is probably metav1.Status, `FromObject` can deal with anything
				return apierr.FromObject(event.Object)
			}

			err = packer.DecompressWorkflow(wf)
			if err != nil {
				return err
			}

			// when we re-establish, we want to start at the same place
			opts.ResourceVersion = wf.ResourceVersion

			switch workflow.Status {
			case reporter.WorkflowPending:
				if handleWorkflowPending(workflow, wf, wsr) != nil {
					re.logger.Err(err, "failed to report workflow status")
				}
			case reporter.WorkflowRunning:
				if handleWorkflowRunning(workflow, wf, wsr) != nil {
					re.logger.Err(err, "failed to report workflow status")
				}
			}
			if hasWorkflowFinished(wf) {
				if handleWorkflowFinished(workflow, wf, wsr) != nil {
					re.logger.Err(err, "failed to report workflow status")
				}
				break Loop
			}
		}
	}

	return nil
}

func handleWorkflowPending(workflow *reporter.Workflow, wf *wfv1.Workflow, wsr *reporter.WorkflowStatusReporter) error {
	if hasWorkflowStarted(wf) {
		workflow.Status = reporter.WorkflowRunning
		for _, node := range wf.Status.Nodes {
			tmpl := wf.GetTemplateByName(node.TemplateName)
			if tmpl == nil || !tmpl.IsPodType() {
				continue
			}
			workflow.Steps[node.DisplayName] = &reporter.WorkflowStep{Status: reporter.WorkflowStepPending, Name: node.DisplayName}
		}
		return wsr.Report(reporter.WorkflowRunning, nil)
	}
	return nil
}

func handleWorkflowRunning(workflow *reporter.Workflow, wf *wfv1.Workflow, wsr *reporter.WorkflowStatusReporter) error {
	for _, node := range wf.Status.Nodes {
		tmpl := wf.GetTemplateByName(node.TemplateName)
		if tmpl == nil || !tmpl.IsPodType() {
			continue
		}
		step, exists := workflow.Steps[node.DisplayName]
		if !exists {
			wsr.Logger.Info("Step status reported but the step does not exits", "step", node.DisplayName)
		}
		if hasStepStatusChanged(step, &node) {
			step.Status = getStepStatus(&node)
			if step.Status == reporter.WorkflowStepFailed {
				if err := wsr.ReportStep(step.Name, reporter.WorkflowStepFailed, getStepError(&node)); err != nil {
					wsr.Logger.Err(err, "failed to report workflow step status")
				}
				continue
			}
			if err := wsr.ReportStep(step.Name, step.Status, nil); err != nil {
				wsr.Logger.Err(err, "failed to report workflow step status")
			}
			continue
		}
	}

	return nil
}

func handleWorkflowFinished(workflow *reporter.Workflow, wf *wfv1.Workflow, wsr *reporter.WorkflowStatusReporter) error {
	if err := hasWorkflowFailed(wf); err != nil {
		return wsr.Report(reporter.WorkflowFailed, err)
	}
	return wsr.Report(reporter.WorkflowSucceded, nil)
}
