# --------------------------
# Build d-operators binary
# --------------------------
FROM golang:1.13.5 as builder

WORKDIR /tools

# copy go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# copy build manifests
COPY Makefile Makefile

# copy source files
COPY pkg/ pkg/
COPY main.go main.go

# build binary
RUN make

# ---------------------------
# Use alpine as minimal base image to package the final binary
# ---------------------------
FROM alpine:latest

WORKDIR /

COPY --from=builder /tools/daction /usr/bin/

CMD ["/usr/bin/daction"]