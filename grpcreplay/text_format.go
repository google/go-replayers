// Copyright 2018 Google LLC
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

package grpcreplay

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	pb "github.com/google/go-replayers/grpcreplay/proto/grpcreplay"
	"google.golang.org/protobuf/encoding/prototext"
)

// Text format:
// First line: magic value (consumed by newReader)
// Second line: initial state, as a Go quoted string
// Each entry: a line with the number of bytes in the marshaled entry,
// followed by the entry.

type textWriter struct {
	w io.Writer
}

func (w *textWriter) writeMagic() error {
	_, err := io.WriteString(w.w, textMagic)
	return err
}

func (w *textWriter) writeHeader(initial []byte) error {
	// initial newline after magic
	_, err := fmt.Fprintf(w.w, "\n%q\n", initial)
	return err
}

func (w *textWriter) writeEntry(e *entry) error {
	pe, err := e.toProto()
	if err != nil {
		return err
	}
	bytes, err := prototext.MarshalOptions{Multiline: true}.Marshal(pe)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w.w, len(bytes)); err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, string(bytes))
	return err
}

type textReader struct {
	r *bufio.Reader
}

func newTextReader(r io.Reader) *textReader {
	return &textReader{bufio.NewReader(r)}
}

func (r *textReader) readHeader() ([]byte, error) {
	// There should a newline after reading magic.
	if _, err := r.readLine(); err != nil {
		return nil, err
	}
	quoted, err := r.readLine()
	if err != nil {
		return nil, err
	}
	if len(quoted) == 0 {
		// just the newline; so no initial state
		return nil, nil
	}
	unquoted, err := strconv.Unquote(string(quoted))
	if err != nil {
		return nil, fmt.Errorf("strconv.Unquote: %w", err)
	}
	return []byte(unquoted), nil
}

// readEntry reads one entry from the replay file r.
// At end of file, it returns (nil, nil).
func (r *textReader) readEntry() (*entry, error) {
	// Read length.
	line, err := r.readLine()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(string(line))
	if err != nil {
		return nil, err
	}
	buf := make([]byte, n)
	if _, err := r.r.Read(buf); err != nil {
		return nil, err
	}
	var pe pb.Entry
	if err := prototext.Unmarshal(buf, &pe); err != nil {
		return nil, err
	}
	return protoToEntry(&pe)
}

// readLine reads a line and returns it without the final newline.
func (r *textReader) readLine() ([]byte, error) {
	line, err := r.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line[:len(line)-1], nil
}
