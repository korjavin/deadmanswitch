#!/bin/bash

# Build and run the new server
go build -o bin/server_new ./cmd/server_new
./bin/server_new
