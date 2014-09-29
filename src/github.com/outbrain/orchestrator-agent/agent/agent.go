/*
   Copyright 2014 Outbrain Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package agent

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/outbrain/log"
	"github.com/outbrain/orchestrator-agent/config"
	"github.com/outbrain/orchestrator-agent/osagent"
)

// httpGet is a convenience method for getting http response from URL, optionaly skipping SSL cert verification
func httpGet(url string) (resp *http.Response, err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.Config.SSLSkipVerify},
	}
	client := &http.Client{Transport: tr}
	return client.Get(url)
}

func SubmitAgent() error {
	hostname, err := osagent.Hostname()
	if err != nil {
		return log.Errore(err)
	}

	url := fmt.Sprintf("%s/api/submit-agent/%s/%d/%s", config.Config.AgentsServer, hostname, config.Config.HTTPPort, ProcessToken.Hash)
	log.Debugf("Submitting this agent: %s", url)

	response, err := httpGet(url)
	if err != nil {
		return log.Errore(err)
	}

	log.Debugf("response: %+v", response)
	return err
}

// ContinuousOperation starts an asynchronuous infinite operation process where:
// - agent is submitted into orchestrator
func ContinuousOperation() {
	log.Infof("Starting continuous operation")
	tick := time.Tick(time.Duration(config.Config.ContinuousPollSeconds) * time.Second)
	resubmitTick := time.Tick(time.Duration(config.Config.ResubmitAgentIntervalMinutes) * time.Minute)

	SubmitAgent()
	for _ = range tick {
		// Do stuff

		// See if we should also forget instances/agents (lower frequency)
		select {
		case <-resubmitTick:
			SubmitAgent()
		default:
		}
	}
}
