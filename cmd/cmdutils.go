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
	b64 "encoding/base64"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"github.com/codefresh-io/status-reporter/pkg/codefresh"
	"github.com/codefresh-io/status-reporter/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultCodefreshHost = "https://g.codefresh.io"
)

var (
	exit = os.Exit
)

func dieOnError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func buildCodefreshClient(host string, token string, httpClient *http.Client, lgr logger.Logger) *codefresh.Codefresh {
	httpHeaders := http.Header{}
	{
		httpHeaders.Add("User-Agent", fmt.Sprintf("codefresh-status-reporter-%s", version))
	}
	return &codefresh.Codefresh{
		Host:       host,
		Token:      token,
		Logger:     lgr,
		HTTPClient: httpClient,
		Headers:    httpHeaders,
	}
}

func buildHTTPClient(rejectTLSUnauthorized bool) *http.Client {
	var httpClient http.Client
	if !rejectTLSUnauthorized {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		httpClient = http.Client{
			Transport: customTransport,
		}
	}

	return &httpClient
}


func BuildKubeClient(host string, token string, b64crt string) (*kubernetes.Clientset, error) {
	ca, err := b64.StdEncoding.DecodeString(b64crt)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(&rest.Config{
		Host:        host,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	})
}