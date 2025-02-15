FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o url-shortnr ./cmd

FROM alpine:latest
WORKDIR /root/

RUN apk add --no-cache libc6-compat

COPY --from=builder app/url-shortnr .
CMD ["./url-shortnr"]
