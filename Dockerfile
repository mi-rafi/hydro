FROM golang:1.20 AS builder
WORKDIR /go/src/github.com/kara/hydro

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -installsuffix cgo -o app github.com/kara/hydro/cmd/hydro

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=0 /go/src/github.com/kara/hydro .
CMD mkdir /var/data
ENV ST_FILE /var/data/time
ENV LISTEN 0.0.0.0:9000
ENTRYPOINT ["/app/app"]