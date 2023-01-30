/*
   Copyright 2022 Docker Compose CLI authors

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

package metrics

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Reporter reports metric events generated by the client.
type Reporter interface {
	Heartbeat(cmd Command)
}

// HTTPReporter reports metric events to an HTTP endpoint.
type HTTPReporter struct {
	client *http.Client
}

// NewHTTPReporter creates a new reporter that will report metric events using
// the provided HTTP client.
func NewHTTPReporter(client *http.Client) HTTPReporter {
	return HTTPReporter{client: client}
}

// Heartbeat reports a metric for aggregation.
func (l HTTPReporter) Heartbeat(cmd Command) {
	entry, err := json.Marshal(cmd)
	if err != nil {
		// impossible: cannot fail on controlled input (i.e. no cycles)
		return
	}

	resp, _ := l.client.Post(
		"http://localhost/usage",
		"application/json",
		bytes.NewReader(entry),
	)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// WriterReporter reports metrics as JSON lines to the provided writer.
type WriterReporter struct {
	w io.Writer
}

// NewWriterReporter creates a new reporter that will write metrics to the
// provided writer as JSON lines.
func NewWriterReporter(w io.Writer) WriterReporter {
	return WriterReporter{w: w}
}

// Heartbeat reports a metric for aggregation.
func (w WriterReporter) Heartbeat(cmd Command) {
	entry, err := json.Marshal(cmd)
	if err != nil {
		// impossible: cannot fail on controlled input (i.e. no cycles)
		return
	}
	entry = append(entry, '\n')
	_, _ = w.w.Write(entry)
}

// MuxReporter wraps multiple reporter instances and reports metrics to each
// instance on invocation.
type MuxReporter struct {
	reporters []Reporter
}

// NewMuxReporter creates a reporter that will report metrics to each of the
// provided reporter instances.
func NewMuxReporter(reporters ...Reporter) MuxReporter {
	return MuxReporter{reporters: reporters}
}

// Heartbeat reports a metric for aggregation.
func (m MuxReporter) Heartbeat(cmd Command) {
	for i := range m.reporters {
		m.reporters[i].Heartbeat(cmd)
	}
}
