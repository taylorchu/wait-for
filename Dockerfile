FROM golang:1.19-alpine AS build

ARG VERSION=latest
ENV CGO_ENABLED=0
RUN go install github.com/taylorchu/wait-for@$VERSION

FROM alpine

COPY --from=build /go/bin/wait-for /usr/bin/
RUN ! ldd /usr/bin/wait-for
