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

package httpreplay_test

// This test requires a Google Cloud project ID.
// See the flag below.

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-replayers/httpreplay"
	googlec "github.com/google/go-replayers/httpreplay/google"
	googleo "golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var projectID = flag.String("project", "", "Google Cloud project ID")

func TestIntegration_RecordAndReplay(t *testing.T) {
	if *projectID == "" {
		t.Skip("missing -project; integration tests skipped")
	}
	httpreplay.DebugHeaders()
	replayFilename := tempFilename(t, "RecordAndReplay*.replay")
	defer os.Remove(replayFilename)
	ctx := context.Background()

	// Record.
	initial := time.Now()
	ibytes, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := httpreplay.NewRecorder(replayFilename, ibytes)
	if err != nil {
		t.Fatal(err)
	}
	ts, err := googleo.DefaultTokenSource(ctx, storage.ScopeFullControl)
	if err != nil {
		t.Fatal(err)
	}
	hc, err := googlec.RecordClient(ctx, rec.Client(), option.WithTokenSource(ts))
	if err != nil {
		t.Fatal(err)
	}
	wanta, wantc := run(t, hc)
	testReadCRC(t, hc, "recording")
	if err := rec.Close(); err != nil {
		t.Fatalf("rec.Close: %v", err)
	}

	// Replay.
	rep, err := httpreplay.NewReplayer(replayFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer rep.Close()
	hc = rep.Client()
	gota, gotc := run(t, hc)
	testReadCRC(t, hc, "replaying")

	if diff := cmp.Diff(gota, wanta); diff != "" {
		t.Error(diff)
	}
	if !bytes.Equal(gotc, wantc) {
		t.Errorf("got %q, want %q", gotc, wantc)
	}
	var gotInitial time.Time
	if err := json.Unmarshal(rep.Initial(), &gotInitial); err != nil {
		t.Fatal(err)
	}
	if !gotInitial.Equal(initial) {
		t.Errorf("initial: got %v, want %v", gotInitial, initial)
	}
}

// TODO(jba): test errors

func run(t *testing.T, hc *http.Client) (*storage.BucketAttrs, []byte) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	b := client.Bucket(*projectID)
	attrs, err := b.Attrs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	obj := b.Object("replay-test")
	w := obj.NewWriter(ctx)
	data := []byte{150, 151, 152}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return attrs, contents
}

func testReadCRC(t *testing.T, hc *http.Client, mode string) {
	const (
		// This is an uncompressed file.
		// See https://cloud.google.com/storage/docs/public-datasets/landsat
		uncompressedBucket = "gcp-public-data-landsat"
		uncompressedObject = "LC08/PRE/044/034/LC80440342016259LGN00/LC80440342016259LGN00_MTL.txt"

		gzippedBucket = "storage-library-test-bucket"
		gzippedObject = "gzipped-text.txt"
	)
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatalf("%s: %v", mode, err)
	}
	defer client.Close()

	uncompressedObj := client.Bucket(uncompressedBucket).Object(uncompressedObject)
	gzippedObj := client.Bucket(gzippedBucket).Object(gzippedObject)

	for _, test := range []struct {
		desc           string
		obj            *storage.ObjectHandle
		offset, length int64
		readCompressed bool // don't decompress a gzipped file

		wantErr bool
		wantLen int // length of contents
	}{
		{
			desc:           "uncompressed, entire file",
			obj:            uncompressedObj,
			offset:         0,
			length:         -1,
			readCompressed: false,
			wantLen:        7903,
		},
		{
			desc:           "uncompressed, entire file, don't decompress",
			obj:            uncompressedObj,
			offset:         0,
			length:         -1,
			readCompressed: true,
			wantLen:        7903,
		},
		{
			desc:           "uncompressed, suffix",
			obj:            uncompressedObj,
			offset:         3,
			length:         -1,
			readCompressed: false,
			wantLen:        7900,
		},
		{
			desc:           "uncompressed, prefix",
			obj:            uncompressedObj,
			offset:         0,
			length:         18,
			readCompressed: false,
			wantLen:        18,
		},
		{
			// When a gzipped file is unzipped by GCS, we can't verify the checksum
			// because it was computed against the zipped contents. There is no
			// header that indicates that a gzipped file is being served unzipped.
			// But our CRC check only happens if there is a Content-Length header,
			// and that header is absent for this read.
			desc:           "compressed, entire file, server unzips",
			obj:            gzippedObj,
			offset:         0,
			length:         -1,
			readCompressed: false,
			wantLen:        11,
		},
		{
			// When we read a gzipped file uncompressed, it's like reading a regular file:
			// the served content and the CRC match.
			desc:           "compressed, entire file, read compressed",
			obj:            gzippedObj,
			offset:         0,
			length:         -1,
			readCompressed: true,
			wantLen:        31,
		},
		{
			desc:           "compressed, partial, read compressed",
			obj:            gzippedObj,
			offset:         1,
			length:         8,
			readCompressed: true,
			wantLen:        8,
		},
		{
			desc:    "uncompressed, HEAD",
			obj:     uncompressedObj,
			offset:  0,
			length:  0,
			wantLen: 0,
		},
		{
			desc:    "compressed, HEAD",
			obj:     gzippedObj,
			offset:  0,
			length:  0,
			wantLen: 0,
		},
	} {
		obj := test.obj.ReadCompressed(test.readCompressed)
		r, err := obj.NewRangeReader(ctx, test.offset, test.length)
		if err != nil {
			if test.wantErr {
				continue
			}
			t.Errorf("%s: %s: %v", mode, test.desc, err)
			continue
		}
		data, err := ioutil.ReadAll(r)
		_ = r.Close()
		if err != nil {
			t.Errorf("%s: %s: %v", mode, test.desc, err)
			continue
		}
		if got, want := len(data), test.wantLen; got != want {
			t.Errorf("%s: %s: len: got %d, want %d", mode, test.desc, got, want)
		}
	}
}

func TestRemoveAndClear(t *testing.T) {
	// Disable logging for this test, since it generates a lot.
	log.SetOutput(ioutil.Discard)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "LGTM")
	}))
	defer srv.Close()

	replayFilename := tempFilename(t, "TestRemoveAndClear*.replay")
	defer os.Remove(replayFilename)

	// Record
	rec, err := httpreplay.NewRecorder(replayFilename, nil)
	if err != nil {
		t.Fatal(err)
	}
	rec.ClearHeaders("Clear")
	rec.RemoveRequestHeaders("Rem*")
	rec.ClearQueryParams("c")
	rec.RemoveQueryParams("r")
	rec.ScrubBody("<secret>.*</secret>")
	hc := rec.Client()
	query := "k=1&r=2&c=3"
	body := "body body <secret>secret</secret> body"
	req, err := http.NewRequest("GET", srv.URL+"?"+query, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	headers := map[string]string{"Keep": "ok", "Clear": "secret", "Remove": "bye"}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if _, err := hc.Do(req); err != nil {
		t.Fatal(err)
	}
	if err := rec.Close(); err != nil {
		t.Fatal(err)
	}

	// Replay
	// For both headers and query param:
	// - k or Keep must be present and identical
	// - c or Clear must be present, but can be different
	// - r or Remove can be anything
	// For body, anything can be between the <secret> and </secret>.
	for _, test := range []struct {
		query       string
		headers     map[string]string
		body        string
		wantSuccess bool
	}{
		{query, headers, body, true}, // same query string, headers, and body
		{query,
			map[string]string{"Keep": "oops", "Clear": "secret", "Remove": "bye"},
			body,
			false, // different Keep
		},
		{query, map[string]string{}, body, false},                               // missing Keep and Clear
		{query, map[string]string{"Keep": "ok"}, body, false},                   // missing Clear
		{query, map[string]string{"Keep": "ok", "Clear": "secret"}, body, true}, // missing Remove is OK
		{
			query,
			map[string]string{"Keep": "ok", "Clear": "secret", "Remove": "whatev"},
			body,
			true,
		}, // different Remove is OK
		{query, map[string]string{"Keep": "ok", "Clear": "diff"}, body, true}, // different Clear is OK
		{"", headers, body, false},            // no query string
		{"k=x&r=2&c=3", headers, body, false}, // different k
		{"r=2", headers, body, false},         // missing k and c
		{"k=1&r=2", headers, body, false},     // missing c
		{"k=1&c=3", headers, body, true},      // missing r is OK
		{"k=1&r=x&c=3", headers, body, true},  // different r is OK,
		{"k=1&r=2&c=x", headers, body, true},  // different clear is OK
		// Note: keep the length of the secret the same to avoid changing the Content-Length header.
		{query, headers, "body body <secret>foobar</secret> body", true}, // different secret in body is OK
		{query, headers, "foo", false},                                   // different body
	} {
		rep, err := httpreplay.NewReplayer(replayFilename)
		if err != nil {
			t.Fatal(err)
		}
		hc = rep.Client()
		url := srv.URL
		if test.query != "" {
			url += "?" + test.query
		}
		req, err = http.NewRequest("GET", url, strings.NewReader(test.body))
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range test.headers {
			req.Header.Set(k, v)
		}
		resp, err := hc.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Close()
		if (resp.StatusCode == 200) != test.wantSuccess {
			t.Errorf("%q, %v, %q: got %d, wanted success=%t",
				test.query, test.headers, test.body, resp.StatusCode, test.wantSuccess)
		}
	}
}

func tempFilename(t *testing.T, pattern string) string {
	f, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return filename
}
