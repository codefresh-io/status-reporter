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

package codefresh

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/codefresh-io/status-reporter/pkg/logger"
	"github.com/codefresh-io/status-reporter/pkg/reporter"
)

const (
	defaultHost = "https://g.codefresh.io"
)

type (
	// Codefresh API client
	Codefresh struct {
		EventReportingURL string
		Token             string
		Logger            logger.Logger
		HTTPClient        *http.Client
		Headers           http.Header
	}

	workflowEvent struct {
		Action string `json:"action,omitempty"`
		Err    string `json:"error,omitempty"`
		Status string `json:"status,omitempty"`
		Step   string `json:"step,omitempty"`
		Name   string `json:"name,omitempty"`
	}
)

func (c *Codefresh) ReportWorkflowStaus(workflow string, status reporter.WorkflowStatus, workflowErr error) error {
	switch status {
	case reporter.WorkflowRunning:
		if err := c.sendStartEvent(); err != nil {
			c.Logger.Err(err, "failed to report start event")
			return err
		}
		c.Logger.Info("reported workflow start")
	case reporter.WorkflowFailed, reporter.WorkflowSucceded:
		if err := c.sendFinishEvent(workflowErr); err != nil {
			c.Logger.Err(err, "failed to report finish event")
			return err
		}
		c.Logger.Info("reported workflow finished", "workflow", workflow, "status", status, "error", workflowErr)
	}
	return nil
}

func (c *Codefresh) ReportWorkflowStepStaus(workflow string, step string, status reporter.WorkflowStepStatus, stepErr error) error {
	switch status {
	case reporter.WorkflowStepRunning:
		if err := c.startStep(step, string(status), stepErr); err != nil {
			c.Logger.Err(err, "failed to report step start event")
			return err
		}
		c.Logger.Info("reported step start", "step", step)
	default:
		if err := c.sendStepStatus(step, string(status), stepErr); err != nil {
			c.Logger.Err(err, "failed to report step status")
			return err
		}
		c.Logger.Info("reported step status", "step", step, "status", status, "error", stepErr)
	}
	return nil
}

func (c *Codefresh) sendStartEvent() error {
	resp, err := c.sendEvent(workflowEvent{Action: "start"})
	if err != nil {
		return err
	}
	c.Logger.Info(string(resp))
	return nil
}

func (c *Codefresh) sendFinishEvent(workflowErr error) error {
	workflowErrStr := ""
	if workflowErr != nil {
		workflowErrStr = workflowErr.Error()
	}
	resp, err := c.sendEvent(workflowEvent{Action: "finish", Err: workflowErrStr})
	if err != nil {
		return err
	}
	c.Logger.Info(string(resp))
	resp, err = c.sendEvent(workflowEvent{Action: "finish-system"})
	if err != nil {
		return err
	}
	c.Logger.Info(string(resp))
	return nil
}

func (c *Codefresh) startStep(step, status string, stepErr error) error {
	resp, err := c.sendEvent(workflowEvent{Action: "pre-steps-succeeded"})
	if err != nil {
		return err
	}
	resp, err = c.sendEvent(workflowEvent{Action: "new-progress-step", Name: step})
	if err != nil {
		return err
	}
	c.Logger.Info(string(resp))
	return c.sendStepStatus(step, status, stepErr)
}

func (c *Codefresh) sendStepStatus(step, status string, stepErr error) error {
	stepErrStr := ""
	if stepErr != nil {
		stepErrStr = stepErr.Error()
	}
	resp, err := c.sendEvent(workflowEvent{Action: "report-status", Step: step, Status: status, Err: stepErrStr})
	if err != nil {
		return err
	}
	c.Logger.Info(string(resp))
	return nil
}

func (c *Codefresh) buildErrorFromResponse(status int, body []byte) error {
	return Error{
		APIStatusCode: status,
		Message:       string(body),
	}
}

func (c *Codefresh) prepareRequest(method, url string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header = c.Headers.Clone()
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (c *Codefresh) sendEvent(ev workflowEvent) ([]byte, error) {
	body, err := json.Marshal(&ev)
	if err != nil {
		return nil, err
	}
	req, err := c.prepareRequest("POST", c.EventReportingURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, c.buildErrorFromResponse(resp.StatusCode, data)
	}
	return data, nil
}
