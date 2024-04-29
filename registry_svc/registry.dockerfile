# base go image
FROM golang:1.21.2-alpine3.18 as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

WORKDIR /app/bin

RUN CGO_ENABLED=0 go build -o ./registry ../cmd

RUN chmod +x /app/bin/registry


# tiny image
# FROM scratch
FROM alpine:3.18.4

RUN mkdir /app 

COPY --from=builder /app/bin/ /app/bin
COPY --from=builder /app/configs/v1/ /app/configs/v1
COPY --from=builder /app/api/v1/ /app/api/v1
# COPY ./bin/broker /app

CMD ["/app/bin/registry"]

