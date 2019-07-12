// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package google provides a way to obtain an http.Client for Google
// Cloud Platform APIs.
package google

import (
	"context"
	"net/http"

	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// Client converts a client returned from Recorder.Client to one configured for
// Google Cloud Platform APIs.
//
// Provide authentication options like option.WithTokenSource as you normally would,
// or omit them to use Application Default Credentials.
func RecordClient(ctx context.Context, c *http.Client, opts ...option.ClientOption) (*http.Client, error) {
	trans, err := htransport.NewTransport(ctx, c.Transport, opts...)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: trans}, nil
}
