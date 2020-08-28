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

package grpcreplay_test

import (
	"github.com/getoutreach/go-replayers/grpcreplay"
	"google.golang.org/grpc"
)

var serverAddress string

func Example_NewRecorder() {
	rec, err := grpcreplay.NewRecorder("service.replay", nil)
	if err != nil {
		// TODO: Handle error.
	}
	defer func() {
		if err := rec.Close(); err != nil {
			// TODO: Handle error.
		}
	}()
	conn, err := grpc.Dial(serverAddress, rec.DialOptions()...)
	if err != nil {
		// TODO: Handle error.
	}
	_ = conn // TODO: use connection
}

func Example_NewReplayer() {
	rep, err := grpcreplay.NewReplayer("service.replay", nil)
	if err != nil {
		// TODO: Handle error.
	}
	defer rep.Close()
	conn, err := rep.Connection()
	if err != nil {
		// TODO: Handle error.
	}
	_ = conn // TODO: use connection
}
