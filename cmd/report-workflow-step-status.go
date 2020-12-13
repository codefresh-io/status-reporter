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
	"os"

	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type reportWorkflowStepCmdOptions struct {
	codefreshToken        string
	codefreshHost         string
	clusterURL            string
	clusterCert           string
	clusterToken          string
	clusterNamespace      string
	workflowID            string
	step                  string
	verbose               bool
	rejectTLSUnauthorized bool
	serverPort            string
}

var (
	reportWorkflowStepOptions reportWorkflowStepCmdOptions
)

var reportWorkflowStepCmd = &cobra.Command{
	Use: "step",

	Run: func(cmd *cobra.Command, args []string) {
		reportWorkflowStepStatus(reportWorkflowStepOptions)
	},
	Long: "Watches and reports workflow steps statusses",
}

func init() {
	dieOnError(viper.BindEnv("codefresh-token", "CODEFRESH_TOKEN"))
	dieOnError(viper.BindEnv("codefresh-host", "CODEFRESH_HOST"))
	dieOnError(viper.BindEnv("workflow", "WORKFLOW_ID"))
	dieOnError(viper.BindEnv("cluster-url", "CLUSTER_URL"))
	dieOnError(viper.BindEnv("cluster-token", "CLUSTER_TOKEN"))
	dieOnError(viper.BindEnv("cluster-namespace", "CLUSTER_NAMESPACE"))
	dieOnError(viper.BindEnv("cluster-cert", "CLUSTER_CERT"))
	dieOnError(viper.BindEnv("tls-reject-unauthorized", "NODE_TLS_REJECT_UNAUTHORIZED"))

	viper.SetDefault("codefresh-host", defaultCodefreshHost)
	viper.SetDefault("port", "8080")
	viper.SetDefault("NODE_TLS_REJECT_UNAUTHORIZED", "1")

	reportWorkflowStepCmd.Flags().BoolVar(&reportWorkflowStepOptions.verbose, "verbose", viper.GetBool("verbose"), "Show more logs")
	reportWorkflowStepCmd.Flags().BoolVar(&reportWorkflowStepOptions.rejectTLSUnauthorized, "tls-reject-unauthorized", viper.GetBool("NODE_TLS_REJECT_UNAUTHORIZED"), "Disable certificate validation for TLS connections")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.codefreshHost, "codefresh-host", viper.GetString("codefresh-host"), "Codefresh API host default [$CODEFRESH_HOST]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.clusterURL, "cluster-url", viper.GetString("cluster-url"), "API URL of the Kubernetes cluster [$CLUSTER_URL]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.clusterToken, "cluster-token", viper.GetString("cluster-token"), "Kubernetes auth token [$CLUSTER_TOKEN]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.clusterNamespace, "cluster-namespace", viper.GetString("cluster-namespace"), "Kubernetes namespace where the workflow is running [$CLUSTER_NAMESPACE]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.clusterCert, "cluster-cert", viper.GetString("cluster-cert"), "Signed certificated authority (base64 encoded) [$CLUSTER_CERT]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.workflowID, "workflow", viper.GetString("workflow"), "Workflow ID to report the status [$WORKFLOW_ID]")

	reportWorkflowStepCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			dieOnError(reportWorkflowStepCmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})

	dieOnError(reportWorkflowStepCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(reportWorkflowStepCmd.MarkFlagRequired("workflow"))
	// dieOnError(reportWorkflowStepCmd.MarkFlagRequired("step"))

	rootCmd.AddCommand(reportWorkflowStepCmd)
}

func reportWorkflowStepStatus(options reportWorkflowStepCmdOptions) {
	log := logger.New(logger.Options{})

	log.Info("Starting watcher", "pid", os.Getpid(), "version", version)
	if !options.rejectTLSUnauthorized {
		log.Info("Running in insecure mode", "NODE_TLS_REJECT_UNAUTHORIZED", options.rejectTLSUnauthorized)
	}

	httpCleint := buildHTTPClient(options.rejectTLSUnauthorized)
	cf := buildCodefreshClient(options.codefreshHost, options.codefreshToken, httpCleint, log)
	kclient, err := BuildKubeClient(options.clusterURL, options.clusterToken, options.clusterCert)
	dieOnError(err)
	res, err := kclient.CoreV1().Pods(options.clusterNamespace).List(metav1.ListOptions{})
	dieOnError(err)
	log.Info("Found pods", "number", len(res.Items))
	stream, err := kclient.CoreV1().Events(options.clusterNamespace).Watch(metav1.ListOptions{
		// Request the API server only events for specific workflow
		// LabelSelector: "",
	})
	dieOnError(err)
	evChannel := stream.ResultChan()
	for {
		if evChannel == nil {
			break
		}
		select {
		case obj, ok := <-evChannel:
			if !ok {
				log.Info("Event channel is closed")
				continue
			}
			ev := obj.Object.(*corev1.Event)
			log.Info("Got event", "source", ev.Source.Component)
			if ev.Source.Component != "workflow-controller" {
				continue
			}

			if ev.InvolvedObject.Kind != "workflow" {
				continue
			}

			// step finished
			if ev.Reason == "WorkflowNodeSucceeded" {
				wssr := reporter.WorkflowStepStatusReporter{
					CodefreshAPI: cf,
					Logger:       log,
					WorkflowID:   options.workflowID,
					Step:         options.step,
				}
				wssr.Report(reporter.WorkflowStepSucceded)
			}

		}
	}
}
