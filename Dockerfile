FROM golang:latest AS builder
WORKDIR /license-proxyserver-addon
COPY . .
ENV GO_PACKAGE github.com/RokibulHasan7/license-proxyserver-addon

# Build
RUN make build --warn-undefined-variables

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:latest

# Add the binaries
WORKDIR /addon/
COPY --from=builder /license-proxyserver-addon/bin/ .
CMD ["./license-proxyserver-addon", "manager"]