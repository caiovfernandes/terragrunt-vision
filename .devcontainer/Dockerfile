# Base image with Go
FROM golang:1.23

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    && apt-get clean

# Install AWS CLI v2
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" \
    && unzip awscliv2.zip \
    && ./aws/install \
    && rm -rf awscliv2.zip ./aws

# Set up Go environment
ENV GO111MODULE=on

