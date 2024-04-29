# base go image
FROM golang:1.21.6-alpine3.19 as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

ENV DOCKER_CONTAINER=true

# RUN go mod tidy
# RUN go mod vendor

# RUN CGO_ENABLED=0 go build -o ./loader ../cmd/loader/.
# RUN CGO_ENABLED=0 go build -o ./broker ../cmd/broker/.

# RUN chmod +x /app/bin/loader
# RUN chmod +x /app/bin/broker

CMD ["/bin/sh", "/app/runDev.sh"]




