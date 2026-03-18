FROM debian:trixie-backports

ARG VERSION
ENV VERSION=${VERSION}

# Avoid interactive prompts during apt installs
ENV DEBIAN_FRONTEND=noninterative

# Install certificates CA
RUN apt-get update \
    && apt-get install -y ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create directory for binary
RUN mkdir -p /go/bin
ENV PATH="/go/bin:${PATH}"

# ---- COPY pre-built binary (CI/CD build job) ----
COPY .bin/wis2-ingest /go/bin/wis2-ingest

# ---- Execution permissions & record version inside container ----
RUN chmod +x /go/bin/wis2-ingest \
    && echo "Docker Image Version: ${VERSION}" > /opt/wis2-ingest.rev

# ---- Entrypoint command ----
ENTRYPOINT ["wis2-ingest"]