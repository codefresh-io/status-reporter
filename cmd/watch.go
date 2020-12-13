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
	"os"

	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var watchWorkflowCmdOptions struct {
	codefreshToken    string
	eventReportingURL string
	clusterNamespace  string
	configPath        string
	contextName       string
	workflowID        string
	runtimeType       string
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
	dieOnError(viper.BindEnv("runtime-type", "RUNTIME_TYPE"))

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
	watchWorkflowCmd.Flags().StringVar(&watchWorkflowCmdOptions.runtimeType, "runtime-type", viper.GetString("runtime-type"), "The type of the runtime environment [tekton/argo] [$RUNTIME_TYPE]")

	watchWorkflowCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			dieOnError(watchWorkflowCmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})

	dieOnError(watchWorkflowCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(watchWorkflowCmd.MarkFlagRequired("workflow"))
	dieOnError(watchWorkflowCmd.MarkFlagRequired("cluster-namespace"))
	dieOnError(watchWorkflowCmd.MarkFlagRequired("runtime-type"))

	rootCmd.AddCommand(watchWorkflowCmd)
}

func watchWorkflowStatus() {
	log := logger.New(logger.Options{})

	log.Info("Starting watcher", "pid", os.Getpid(), "version", version)

	httpClient := buildHTTPClient(true)
	cf := buildCodefreshClient(watchWorkflowCmdOptions.eventReportingURL, watchWorkflowCmdOptions.codefreshToken, httpClient, log)
	cnf, err := BuildRestConfig(watchWorkflowCmdOptions.configPath, watchWorkflowCmdOptions.contextName, watchWorkflowCmdOptions.inCluster, log)
	dieOnError(err)

	re, err := RuntimeFactory(watchWorkflowCmdOptions.runtimeType, &runtime.Options{
		Logger: log,
		Client: cf,
		Config: cnf,
	})
	dieOnError(err)

	err = re.Watch(context.TODO(), watchWorkflowCmdOptions.clusterNamespace, reportWorkflowOptions.workflowID)
}
