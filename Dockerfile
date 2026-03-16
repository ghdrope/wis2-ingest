FROM debian:trixie-backports

# Avoid interactive prompts during apt installs
ENV DEBIAN_FRONTEND=noninterative

# Create directory for binary
RUN mkdir -p /go/bin
ENV PATH="/go/bin:${PATH}"

# ---- COPY pre-built binary (CI/CD build job) ----
COPY .bin/wis2-ingest /go/bin/wis2-ingest

# ---- Execution permissions & record version inside container ----
RUN chmod +x /go/bin/wis2-ingest \
    && echo ${VERSION} > /opt/wis2-ingest.rev

# ---- Entrypoint command ----
ENTRYPOINT ["wis2-ingest"]