#!/bin/bash

cd ./bin
go build -o ./loader ../cmd/loader/.
./loader

go build -o ./broker ../cmd/broker/.
./broker



