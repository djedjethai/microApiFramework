CERTIFS_PATH=$(shell pwd)/external-storage/certifs

AUTH_BINARY=auth
BROKER_BINARY=broker

.PHONY: up_build
up_build: build_auth build_broker
	@echo "Stopping docker images (if running...)"
	sudo docker-compose down
	@echo "Building (when required) and starting docker images..."
	sudo docker-compose up --build
	@echo "Docker images built and started!" 

.PHONY: up_build_d
up_build_d: build_auth build_broker
	@echo "Stopping docker images (if running...)"
	sudo docker-compose down
	@echo "Building (when required) and starting docker images..."
	sudo docker-compose up --build -d
	@echo "Docker images built and started!" 

.PHONY: build_auth
build_auth:
	@echo "Building auth binary..."
	cd ./auth-svc && env GOOS=linux CGO_ENABLED=0 go build -o bin/${AUTH_BINARY} ./cmd
	@echo "Done!"

.PHONY: build_broker
build_broker:
	@echo "Building broker binary..."
	cd ./broker-svc/broker && env GOOS=linux CGO_ENABLED=0 go build -o bin/${BROKER_BINARY} ./cmd
	@echo "Done!"


.PHONY: gencert
gencert:
	cfssl gencert \
		-initca test/ca-csr.json | cfssljson -bare ca
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		test/client-csr.json | cfssljson -bare client	
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server \
		test/server-csr.json | cfssljson -bare server
	mv *.pem *.csr ${CERTIFS_PATH}


