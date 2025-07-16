FROM --platform=$BUILDPLATFORM golang:1.24-alpine@sha256:ddf52008bce1be455fe2b22d780b6693259aaf97b16383b6372f4b22dd33ad66 as builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /diffbot

COPY . .
# go env GOCACHE; go env GOMODCACHE
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/diffbot .

FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

ARG TARGETOS
ARG TARGETARCH
ADD --chmod=0555 https://github.com/argoproj/argo-cd/releases/download/v3.0.11/argocd-${TARGETOS}-${TARGETARCH} /usr/local/bin/argocd
COPY --from=builder /bin/diffbot /usr/local/bin/
