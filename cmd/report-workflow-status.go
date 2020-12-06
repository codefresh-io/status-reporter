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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/codefresh-io/status-reporter/pkg/codefresh"
	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type reportWorkflowCmdOptions struct {
	codefreshToken        string
	codefreshHost         string
	workflowID            string
	verbose               bool
	rejectTLSUnauthorized bool
	serverPort            string
}

var (
	reportWorkflowOptions reportWorkflowCmdOptions
)

var reportWorkflowCmd = &cobra.Command{
	Use: "workflow",

	Run: func(cmd *cobra.Command, args []string) {
		reportWorkflowStatus(reportWorkflowOptions)
	},
	Long: "Report workflow status",
}

func init() {
	dieOnError(viper.BindEnv("codefresh-token", "CODEFRESH_TOKEN"))
	dieOnError(viper.BindEnv("codefresh-host", "CODEFRESH_HOST"))
	dieOnError(viper.BindEnv("workflow", "WORKFLOW_ID"))
	dieOnError(viper.BindEnv("NODE_TLS_REJECT_UNAUTHORIZED"))

	viper.SetDefault("codefresh-host", defaultCodefreshHost)
	viper.SetDefault("port", "8080")
	viper.SetDefault("NODE_TLS_REJECT_UNAUTHORIZED", "1")

	reportWorkflowCmd.Flags().BoolVar(&reportWorkflowOptions.verbose, "verbose", viper.GetBool("verbose"), "Show more logs")
	reportWorkflowCmd.Flags().BoolVar(&reportWorkflowOptions.rejectTLSUnauthorized, "tls-reject-unauthorized", viper.GetBool("NODE_TLS_REJECT_UNAUTHORIZED"), "Disable certificate validation for TLS connections")
	reportWorkflowCmd.Flags().StringVar(&reportWorkflowOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	reportWorkflowCmd.Flags().StringVar(&reportWorkflowOptions.codefreshHost, "codefresh-host", viper.GetString("codefresh-host"), "Codefresh API host default [$CODEFRESH_HOST]")
	reportWorkflowCmd.Flags().StringVar(&reportWorkflowOptions.workflowID, "workflow", viper.GetString("workflow"), "Workflow ID to report the status [$WORKFLOW_ID]")

	reportWorkflowCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			dieOnError(reportWorkflowCmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})

	dieOnError(reportWorkflowCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(reportWorkflowCmd.MarkFlagRequired("workflow"))

	rootCmd.AddCommand(reportWorkflowCmd)
}

func reportWorkflowStatus(options reportWorkflowCmdOptions) {
	log := logger.New(logger.Options{})

	log.Info("Starting", "pid", os.Getpid(), "version", version)
	if !options.rejectTLSUnauthorized {
		log.Info("Running in insecure mode", "NODE_TLS_REJECT_UNAUTHORIZED", options.rejectTLSUnauthorized)
	}

	var cf *codefresh.Codefresh
	{
		var httpClient http.Client
		if !options.rejectTLSUnauthorized {
			customTransport := http.DefaultTransport.(*http.Transport).Clone()
			// #nosec
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			httpClient = http.Client{
				Transport: customTransport,
			}
		}

		httpHeaders := http.Header{}
		{
			httpHeaders.Add("User-Agent", fmt.Sprintf("codefresh-runner-%s", version))
		}
		cf = &codefresh.Codefresh{
			Host:       options.codefreshHost,
			Token:      options.codefreshToken,
			Logger:     log.Fork("module", "service", "service", "codefresh"),
			HTTPClient: &httpClient,
			Headers:    httpHeaders,
		}
	}

	wsr := reporter.WorkflowStatusReporter{
		CodefreshAPI: cf,
		Logger:       log,
		WorkflowID:   options.workflowID,
	}
	dieOnError(wsr.Report(reporter.WorkflowFailed))
}
