ARG GO_VERSION=1.23.2

FROM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Install upx
RUN apk add upx

# Set the Current Working Directory inside the container
WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN GOPROXY=https://goproxy.io,direct go mod download

COPY . ./

# Enables module-aware mode, regardless of whether the project is inside or outside GOPATH.
ENV GO111MODULE=on
# Enable CGO and build the Go application
ENV CGO_ENABLED=1

# The -ldflags "-s -w" flags to disable the symbol table and DWARF generation that is supposed to create debugging data
RUN go build -ldflags "-s -w" -v -o edge ./cmd/main.go
RUN upx -9 /app/edge


# Start fresh from a smaller image
FROM alpine:latest
RUN apk add ca-certificates

# Install SQLite
RUN apk add --no-cache sqlite

COPY --from=builder /app/edge /app/edge
COPY --from=builder /app/configs/application-dev.yml /app/configs/application-dev.yml
COPY --from=builder /app/configs/application-prod.yml /app/configs/application-prod.yml
COPY --from=builder /app/configs/application-test.yml /app/configs/application-test.yml
COPY --from=builder /app/logs/edge-app.log /app/logs/edge-app.log
COPY --from=builder /app/configs/banner.txt /app/configs/banner.txt


# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the binary program produced by `go install`
CMD ["/app/edge"]