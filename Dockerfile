# Stage 1: Build stage
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy project code
COPY . .

# Set environment variables, CGO_ENABLED=0 to generate statically linked executable
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Build the project
RUN go build -a -installsuffix cgo -o flextopo-agent cmd/agent/main.go

# Stage 2: Runtime stage, using a smaller base image
FROM alpine:3.20

# install lscpu and nvidia-smi
RUN apk update && \
    apk add --no-cache \
    util-linux \
    nvidia-container-runtime && \
    # clean apk cache
    rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /root/

# Copy the compiled executable from the build stage
COPY --from=builder /app/flextopo-agent .

# Create mount points for host paths that need to be accessed
VOLUME [ "/host-sys", "/host-proc" ]

# Expose necessary ports (if needed)
# EXPOSE 8080

# Set the command to run when the container starts
ENTRYPOINT ["./flextopo-agent"]