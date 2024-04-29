#!/bin/bash

cd ./bin
go build -o ./loader ../cmd/loader/.
./loader

go build -o ./order ../cmd/order/.
./order


# cd ./loader
# go build -o loader .
# ./loader
# 
# cd ../order
# go build -o order .
# ./order


