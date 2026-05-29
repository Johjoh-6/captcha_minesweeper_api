#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o=./bin/linux_amd64/api ./cmd/api
