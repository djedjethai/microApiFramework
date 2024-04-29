#!/bin/bash

# Function to perform cleanup
cleanup() {
    echo "Clean up the folders."
    # cd /app

    # Remove specified files and directories
    rm -rf loader
    rm -rf broker/internal
    rm broker/broker.go
    rm broker/setup.go
    rm -rf cmd
    rm -rf vendor
    rm go.sum
    rm go.mod
    rm Makefile
    rm run.sh
    rm broker.dockerfile
}


# Check if the script is running in a Docker environment
if [ -z "$DOCKER_CONTAINER" ]; then
    echo "Error: This script should only be run inside a Docker container."
    exit 1
fi

app_path="/app"
bin_path="/app/bin"

loaderPath=$(find "$bin_path" -type f -name "loader")
brokerPath=$(find "$bin_path" -type f -name "broker")

cd "$bin_path"
ls
if [ -n "$loaderPath" ]; then
	# echo "File loader found in the specified path '$loaderPath'."
	./loader
else
	# echo "File loader NOT found in the specified path '$loaderPath'."
	go build -o ./loader ../cmd/loader/.
	./loader
fi


if [ -n "$brokerPath" ]; then
	# echo "File broker found in the specified path '$brokerPath'."
	./broker
else
	# echo "File broker NOT found in the specified path '$brokerPath'."
	go build -o ./broker ../cmd/broker/.
	cd "$app_path"
	cleanup
	cd "$bin_path"
	./broker
fi

# cd /app/bin
# go build -o ./loader ../cmd/loader/.
# ./loader
# go build -o ./broker ../cmd/broker/.
# ./broker


