FROM golang:1.19-alpine AS build

WORKDIR /go/src/github.com/taylorchu/wait-for/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /tmp .

FROM alpine

COPY --from=build /tmp/wait-for /usr/bin/
RUN ! ldd /usr/bin/wait-for
