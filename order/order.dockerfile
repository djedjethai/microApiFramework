# base go image
FROM golang:1.21.6-alpine3.19 as builder

# RUN apk --no-cache add git

# RUN git config --global --add url."git@gitlab.com:".insteadOf "https://gitlab.com/"

# Needed to be able to build kafka
RUN apk add --no-progress --no-cache gcc musl-dev

RUN mkdir /app

COPY . /app

WORKDIR /app

# ENV GO111MODULE=on

# RUN go get github.com/confluentinc/confluent-kafka-go/v2

# RUN CGO_ENABLED=0 go build -o ./loader ../cmd/loader/.
# RUN CGO_ENABLED=0 go build -o ./broker ../cmd/broker/.

# RUN chmod +x /app/bin/loader
# RUN chmod +x /app/bin/broker

CMD ["/bin/sh", "/app/runDev.sh"]

