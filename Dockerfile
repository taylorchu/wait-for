FROM golang:1.19-alpine AS build

WORKDIR /build/
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ENV CGO_ENABLED=0
RUN go build -o /tmp/wait-for .

FROM alpine

COPY --from=build /tmp/wait-for /usr/bin/
RUN ! ldd /usr/bin/wait-for

CMD ["/usr/bin/wait-for"]
