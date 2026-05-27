FROM debian:trixie-backports

# Must match GitHub repository name
ARG PROJECT_NAME="wis2-ingest"
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
COPY .bin/${PROJECT_NAME} /go/bin/${PROJECT_NAME}

# ---- Execution permissions ----
RUN chmod +x /go/bin/${PROJECT_NAME}
