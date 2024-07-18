# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Makefile for building Go files from protos.

MODULE_PATH=github.com/google/go-replayers/grpcreplay

all: grpcreplay/grpcreplay.pb.go intstore/intstore.pb.go

grpcreplay/grpcreplay.pb.go: proto-prereqs 
	cd grpcreplay; \
	protoc --go_out=. --go-grpc_out=. -I $(PROTOBUF_DIR)/src -I . *.proto && \
	mv $(MODULE_PATH)/proto/grpcreplay/grpcreplay.pb.go . && \
	rmdir -p $(MODULE_PATH)/proto/grpcreplay

intstore/intstore.pb.go: proto-prereqs
	cd intstore; \
	protoc --go_out=. --go-grpc_out=. -I $(PROTOBUF_DIR)/src -I . *.proto && \
	mv $(MODULE_PATH)/proto/intstore/intstore.pb.go . && \
	mv $(MODULE_PATH)/proto/intstore/intstore_grpc.pb.go . && \
	rmdir -p $(MODULE_PATH)/proto/intstore

proto-prereqs: protoc protoc-gen-go protoc-gen-go-grpc $(PROTOBUF_DIR)/src  	

.PHONY: proto-prereqs

protoc:
	@echo $(PATH)
	@if [ -z $$(which protoc) ]; then \
		echo >&2 "add protoc to your path; download it from https://github.com/protocolbuffers/protobuf/releases"; \
		exit 1; \
	fi

protoc-gen-go:
	if [ -z $$(which $@) ]; then \
		go install google.golang.org/protobuf/cmd/$(@)@latest; \
	fi

protoc-gen-go-grpc:
	if [ -z $$(which $@) ]; then \
		go install google.golang.org/grpc/cmd/$(@)@latest; \
	fi
	
$(PROTOBUF_DIR)/src:
	@echo >&2 "download the protobuf release from https://github.com/protocolbuffers/protobuf/releases and set PROTOBUF_DIR to its location"
	@exit 1
