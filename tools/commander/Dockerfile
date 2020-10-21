# --------------------------
# Build d-operators binary
# --------------------------
FROM golang:1.13.5 as builder

WORKDIR /mayadata.io/commander/

COPY . .

# build binary
RUN make

# ---------------------------
# Use distroless as minimal base image to package the final binary
# Refer https://github.com/GoogleContainerTools/distroless
# ---------------------------
FROM alpine:latest

WORKDIR /

COPY --from=builder /mayadata.io/commander/dcmd /usr/bin/

CMD ["/usr/bin/dcmd"]