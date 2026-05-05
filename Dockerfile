# Minimal runtime image. The binary is built by GoReleaser and copied in.
FROM gcr.io/distroless/static-debian12:nonroot

COPY shipnote /usr/local/bin/shipnote

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/shipnote"]
