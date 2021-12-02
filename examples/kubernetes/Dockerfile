FROM golang:1.16-alpine3.12
RUN apk update && apk add --no-cache git curl wget
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/app
ENV EASYRAFT_PORT=5000
ENV HTTP_PORT=5001
ENV DISCOVERY_PORT=5002
ENV DATA_DIR=/data
ENTRYPOINT ["/go/bin/app"]