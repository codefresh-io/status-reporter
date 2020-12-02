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
	"time"

	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type reportWorkflowStepCmdOptions struct {
	codefreshToken        string
	codefreshHost         string
	argoServiceHost       string
	argoServicePort       string
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
	dieOnError(viper.BindEnv("argo-host", "ARGO_SERVER_SERVICE_HOST"))
	dieOnError(viper.BindEnv("argo-port", "ARGO_SERVER_SERVICE_PORT"))
	dieOnError(viper.BindEnv("workflow", "WORKFLOW_ID"))
	dieOnError(viper.BindEnv("step", "STEP"))
	dieOnError(viper.BindEnv("tls-reject-unauthorized", "NODE_TLS_REJECT_UNAUTHORIZED"))

	viper.SetDefault("codefresh-host", defaultCodefreshHost)
	viper.SetDefault("port", "8080")
	viper.SetDefault("NODE_TLS_REJECT_UNAUTHORIZED", "1")

	reportWorkflowStepCmd.Flags().BoolVar(&reportWorkflowStepOptions.verbose, "verbose", viper.GetBool("verbose"), "Show more logs")
	reportWorkflowStepCmd.Flags().BoolVar(&reportWorkflowStepOptions.rejectTLSUnauthorized, "tls-reject-unauthorized", viper.GetBool("NODE_TLS_REJECT_UNAUTHORIZED"), "Disable certificate validation for TLS connections")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.codefreshHost, "codefresh-host", viper.GetString("codefresh-host"), "Codefresh API host default [$CODEFRESH_HOST]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.argoServiceHost, "argo-host", viper.GetString("argo-host"), "Argo host [$ARGO_SERVER_SERVICE_HOST]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.argoServicePort, "argo-port", viper.GetString("argo-port"), "Argo port [$ARGO_SERVER_SERVICE_PORT]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.workflowID, "workflow", viper.GetString("workflow"), "Workflow ID to report the status [$WORKFLOW_ID]")
	reportWorkflowStepCmd.Flags().StringVar(&reportWorkflowStepOptions.step, "step", viper.GetString("step"), "Step name to report the status [$STEP]")

	reportWorkflowStepCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			dieOnError(reportWorkflowStepCmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})

	dieOnError(reportWorkflowStepCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(reportWorkflowStepCmd.MarkFlagRequired("workflow"))
	dieOnError(reportWorkflowStepCmd.MarkFlagRequired("step"))

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
	for {
		wssr := reporter.WorkflowStepStatusReporter{
			CodefreshAPI: cf,
			Logger:       log,
			WorkflowID:   options.workflowID,
			Step:         options.step,
		}
		dieOnError(wssr.Report(reporter.WorkflowFailed))
		time.Sleep(1 * time.Second)
	}
}
