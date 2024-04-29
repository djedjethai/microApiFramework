# base go image
FROM golang:1.21.2-alpine3.18 as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o ./bin/auth ./cmd

RUN chmod +x /app/bin/auth

# tiny image
# FROM scratch
FROM alpine:3.18.4

RUN mkdir /app 

COPY --from=builder /app/bin/auth /app
# COPY /app/bin/auth /app

CMD ["/app/auth"]


