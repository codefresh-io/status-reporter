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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/codefresh-io/status-reporter/pkg/logger"
)

const (
	defaultHost = "https://g.codefresh.io"
)

type (
	// Codefresh API client
	Codefresh struct {
		Host       string
		Token      string
		Logger     logger.Logger
		HTTPClient *http.Client
		Headers    http.Header
	}
)

func (c *Codefresh) ReportWorkflowStaus(workflow string, status string) error {
	return nil
}
func (c *Codefresh) ReportWorkflowStepStaus(workflow string, step string, status string) error {
	return nil
}

func (c *Codefresh) buildErrorFromResponse(status int, body []byte) error {
	return Error{
		APIStatusCode: status,
		Message:       string(body),
	}
}

func (c *Codefresh) prepareURL(paths ...string) (*url.URL, error) {
	if c.Host == "" {
		c.Host = defaultHost
	}
	u, err := url.Parse(c.Host)
	if err != nil {
		return nil, err
	}
	accPath := []string{}
	accRawPath := []string{}

	for _, p := range paths {
		accRawPath = append(accRawPath, url.PathEscape(p))
		accPath = append(accPath, p)
	}
	u.Path = path.Join(accPath...)
	u.RawPath = path.Join(accRawPath...)
	return u, nil
}

func (c *Codefresh) prepareRequest(method string, data io.Reader, apis ...string) (*http.Request, error) {
	u, err := c.prepareURL(apis...)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), data)
	if err != nil {
		return nil, err
	}
	req.Header = c.Headers.Clone()
	if c.Token != "" {
		req.Header.Add("Authorization", c.Token)
	}
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (c *Codefresh) doRequest(method string, body io.Reader, apis ...string) ([]byte, error) {
	req, err := c.prepareRequest(method, body, apis...)
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
