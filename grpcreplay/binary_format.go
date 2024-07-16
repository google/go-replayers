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
	"encoding/binary"
	"errors"
	"io"

	pb "github.com/google/go-replayers/grpcreplay/proto/grpcreplay"
	"google.golang.org/protobuf/proto"
)

type binaryWriter struct {
	w io.Writer
}

func (w *binaryWriter) writeMagic() error {
	_, err := io.WriteString(w.w, binaryMagic)
	return err
}

func (w *binaryWriter) writeHeader(initial []byte) error {
	return w.writeRecord(initial)
}

func (w *binaryWriter) writeEntry(e *entry) error {
	pe, err := e.toProto()
	if err != nil {
		return err
	}
	bytes, err := proto.Marshal(pe)
	if err != nil {
		return err
	}
	return w.writeRecord(bytes)
}

// A record consists of an unsigned 32-bit little-endian length L followed by L
// bytes.

func (w *binaryWriter) writeRecord(data []byte) error {
	if err := binary.Write(w.w, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err := w.w.Write(data)
	return err
}

type binaryReader struct {
	r io.Reader
}

func newBinaryReader(r io.Reader) *binaryReader {
	return &binaryReader{r}
}

func (r *binaryReader) readHeader() ([]byte, error) {
	// Magic has already been read.
	bytes, err := r.readRecord()
	if err == io.EOF {
		err = errors.New("grpcreplay: missing initial state")
	}
	return bytes, err
}

// readEntry reads one entry from the replay file r.
// At end of file, it returns (nil, nil).
func (r *binaryReader) readEntry() (*entry, error) {
	buf, err := r.readRecord()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var pe pb.Entry
	if err := proto.Unmarshal(buf, &pe); err != nil {
		return nil, err
	}
	return protoToEntry(&pe)
}

func (r *binaryReader) readRecord() ([]byte, error) {
	var size uint32
	if err := binary.Read(r.r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
