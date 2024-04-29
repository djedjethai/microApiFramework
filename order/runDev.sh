#!/bin/bash
cd /app/bin
go build -o ./loader ../cmd/loader/.
./loader

# go build -o ./order ../cmd/order/.
# Needed to be able to build kafka
go build -o ./order -tags musl -ldflags '-extldflags "-static"' ../cmd/order/.
./order



