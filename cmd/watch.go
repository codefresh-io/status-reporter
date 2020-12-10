// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
	"github.com/codefresh-io/status-reporter/pkg/tekton"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	tkn "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var watchWorkflowCmdOptions struct {
	codefreshToken    string
	eventReportingURL string
	clusterNamespace  string
	configPath        string
	contextName       string
	workflowID        string
	inCluster         bool
	verbose           bool
}

var watchWorkflowCmd = &cobra.Command{
	Use: "watch",

	Run: func(cmd *cobra.Command, args []string) {
		watchWorkflowStatus()
	},
	Long: "Watches and reports workflow statuses",
}

func init() {
	dieOnError(viper.BindEnv("codefresh-token", "CODEFRESH_TOKEN"))
	dieOnError(viper.BindEnv("codefresh-host", "CODEFRESH_HOST"))
	dieOnError(viper.BindEnv("workflow", "WORKFLOW_ID"))
	dieOnError(viper.BindEnv("cluster-namespace", "CLUSTER_NAMESPACE"))
	dieOnError(viper.BindEnv("config-path", "CONFIG_PATH"))
	dieOnError(viper.BindEnv("context-name", "CONTEXT_NAME"))

	viper.SetDefault("event-reporting-url", defaultCodefreshHost)
	viper.SetDefault("port", "8080")

	watchWorkflowCmd.Flags().BoolVar(&watchWorkflowCmdOptions.verbose, "verbose", viper.GetBool("verbose"), "Show more logs")
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.eventReportingURL, "event-reporting-url", viper.GetString("event-reporting-url"), "Codefresh API host default [$CODEFRESH_HOST]")
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.clusterNamespace, "cluster-namespace", viper.GetString("cluster-namespace"), "Kubernetes namespace where the workflow is running [$CLUSTER_NAMESPACE]")
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.configPath, "config-path", viper.GetString("config-path"), "Kubernetes config path to use [$CONFIG_PATH]")
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.contextName, "context-name", viper.GetString("context-name"), "Kubernetes context name [$CONTEXT_NAME]")
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.workflowID, "workflow", viper.GetString("workflow"), "Workflow ID to report the status [$WORKFLOW_ID]")
	watchWorkflowCmd.Flags().BoolVar(&watchWorkflowCmdOptions.inCluster, "in-cluster", viper.GetBool("in-cluster"), "Should be true if running from inside the cluster")

	watchWorkflowCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			dieOnError(watchWorkflowCmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})

	dieOnError(watchWorkflowCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(watchWorkflowCmd.MarkFlagRequired("workflow"))
	dieOnError(watchWorkflowCmd.MarkFlagRequired("cluster-namespace"))

	rootCmd.AddCommand(watchWorkflowCmd)
}

func watchWorkflowStatus() {
	log := logger.New(logger.Options{})

	log.Info("Starting watcher", "pid", os.Getpid(), "version", version)

	httpClient := buildHTTPClient(true)
	cf := buildCodefreshClient(watchWorkflowCmdOptions.eventReportingURL, watchWorkflowCmdOptions.codefreshToken, httpClient, log)
	tektonClient, err := BuildTektonClient(watchWorkflowCmdOptions.configPath, watchWorkflowCmdOptions.contextName, watchWorkflowCmdOptions.inCluster)
	dieOnError(err)

	watch, err := tektonClient.TektonV1beta1().PipelineRuns(watchWorkflowCmdOptions.clusterNamespace).Watch(context.TODO(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/pipeline=%s", watchWorkflowCmdOptions.workflowID),
		Watch:         true,
	})
	dieOnError(err)

	log.Info("Watching tekton pipelines", "namespace", watchWorkflowCmdOptions.clusterNamespace)

	workflow := reporter.NewWorkflow()
	wsr := &reporter.WorkflowStatusReporter{
		CodefreshAPI: cf,
		Logger:       log,
		WorkflowID:   watchWorkflowCmdOptions.workflowID,
	}

	for ev := range watch.ResultChan() {
		pr, ok := ev.Object.(*tkn.PipelineRun)
		if !ok {
			log.Err(fmt.Errorf("Invalid object type"), "unexpected object type from event")
		}
		switch workflow.Status {
		case reporter.WorkflowPending:
			if handleWorkflowPending(workflow, pr, wsr) != nil {
				log.Err(err, "failed to report workflow status")
			}
		case reporter.WorkflowRunning:
			if handleWorkflowRunning(workflow, pr, wsr) != nil {
				log.Err(err, "failed to report workflow status")
			}
		}
		if tekton.PipelineHasFinished(pr) {
			watch.Stop()
			if handleWorkflowFinished(workflow, pr, wsr) != nil {
				log.Err(err, "failed to report workflow status")
			}
			break
		}
	}

	log.Info("Workflow finished, exiting")
}

func handleWorkflowPending(workflow *reporter.Workflow, pr *tkn.PipelineRun, wsr *reporter.WorkflowStatusReporter) error {
	if tekton.PipelineHasStarted(pr) {
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
		if tekton.HasStepStatusChanged(step, trs) {
			newStatus, err := tekton.GetTaskStatus(trs)
			if err != nil {
				wsr.Logger.Err(err, "failed to get workflow step status")
				continue
			}
			step.Status = newStatus
			if newStatus == reporter.WorkflowStepFailed {
				if err = wsr.ReportStep(step.Name, reporter.WorkflowStepFailed, tekton.TaskHasFailed(trs)); err != nil {
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
	if err := tekton.PipelineHasFailed(pr); err != nil {
		return wsr.Report(reporter.WorkflowFailed, err)
	}
	return wsr.Report(reporter.WorkflowSucceded, nil)
}
