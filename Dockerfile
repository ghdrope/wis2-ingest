FROM golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

# Must match GitHub repository name
ARG PROJECT_NAME="wis2-ingest"
ARG VERSION
ENV VERSION=${VERSION}

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
      -ldflags "\
        -s -w \
        -X github.com/ghdrope/go-version.Version=${VERSION} \
        -X github.com/ghdrope/go-version.GitCommit=${GIT_COMMIT} \
        -X github.com/ghdrope/go-version.BuildDate=${BUILD_DATE}" \
      -o /out/${PROJECT_NAME} \
      ./cmd  



FROM debian:trixie-backports

ARG VERSION
ENV VERSION=${VERSION}

# Avoid interactive prompts during apt installs
ENV DEBIAN_FRONTEND=noninteractive

# Install certificates CA
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/${PROJECT_NAME} /usr/local/bin/${PROJECT_NAME}

# ---- Execution permissions ----
RUN chmod +x /usr/local/bin/${PROJECT_NAME}
