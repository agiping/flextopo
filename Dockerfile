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

# # Stage 2: Runtime stage, using a smaller base image
# FROM alpine:3.20

# # install lscpu and nvidia-smi
# RUN apk update && \
#     apk add --no-cache \
#     util-linux \
#     nvidia-container-runtime && \
#     # clean apk cache
#     rm -rf /var/cache/apk/*
# Stage 2: Runtime stage, using a smaller base image
# FROM ubuntu:22.04

FROM nvidia/cuda:12.0-base

# Install prerequisites
RUN apt-get update && \
    apt-get install -y curl gnupg2 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Add NVIDIA package repositories
# RUN curl -s -L https://nvidia.github.io/nvidia-container-runtime/gpgkey | apt-key add - && \
#     distribution=$(. /etc/os-release;echo $ID$VERSION_ID) && \
#     curl -s -L https://nvidia.github.io/nvidia-container-runtime/$distribution/nvidia-container-runtime.list | \
#     tee /etc/apt/sources.list.d/nvidia-container-runtime.list

# Update and install nvidia-container-runtime
# RUN apt-get update && \
#     apt-get install -y nvidia-container-runtime && \
#     rm -rf /var/lib/apt/lists/*

# Install crictl
RUN VERSION="v1.28.0" && \
    curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$VERSION/crictl-$VERSION-linux-amd64.tar.gz -o /tmp/crictl.tar.gz && \
    tar zxvf /tmp/crictl.tar.gz -C /usr/local/bin && \
    rm -rf /tmp/crictl.tar.gz

# Clean up
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Set working directory
WORKDIR /root/

# Copy the compiled executable from the build stage
COPY --from=builder /app/flextopo-agent .

# Create mount points for host paths that need to be accessed
VOLUME [ "/host-sys", "/host-proc", "/run/containerd/containerd.sock" ]

# Expose necessary ports (if needed)
# EXPOSE 8080

# Set the command to run when the container starts
ENTRYPOINT ["./flextopo-agent"]