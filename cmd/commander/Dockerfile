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
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /mayadata.io/commander/dcmd /usr/bin/

USER nonroot:nonroot

CMD ["/usr/bin/dcmd"]