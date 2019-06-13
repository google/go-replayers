[![Build Status](https://travis-ci.org/google/go-replayers.svg?branch=master)](https://travis-ci.org/google/go-replayers)
[![godoc](https://godoc.org/github.com/google/go-replayers?status.svg)](https://godoc.org/github.com/google/go-replayers)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/go-replayers)](https://goreportcard.com/report/github.com/google/go-replayers)

# Go Tools for Recording and Replaying RPCs

This repo contains two tools for testing network clients by recording real
interactions with servers, then playing back the server responses later. These
record/replay proxies let you run an "integration" test that accesses a backend
and record the interaction. Subsequent runs of the test can replay the server's
responses without actually contacting the server, turning the integration test
into a fast and inexpensive unit test.

To use a record/replay proxy:

1. Write a test that interacts with an actual backend.
2. Run the test using the proxy in record mode. The result will be a file of
   client-server interactions.
3. Run the test again using the proxy in replay mode, pointing at the recorded file.
   The proxy accepts requests and replays the matching server responses.
   

## httpreplay


The httpreplay proxy works with HTTP traffic. You can use the [`httpreplay`
package](https://godoc.org/github.com/google/go-replayers/httpreplay)
directly from Go code, or you can run the `httpr` command at
`httpreplay/cmd/httpr` and use a client written in any language.

## grpcreplay

The [`grpcreplay` package](https://godoc.org/github.com/google/go-replayers/grpcreplay)
works with [gRPC](https://grpc.io) traffic. There is no corresponding command.

