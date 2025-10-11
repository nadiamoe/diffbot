FROM --platform=$BUILDPLATFORM golang:1.25-alpine@sha256:06cdd34bd531b810650e47762c01e025eb9b1c7eadd191553b91c9f2d549fae8 as builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /diffbot

COPY . .
# go env GOCACHE; go env GOMODCACHE
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/diffbot .

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

ARG TARGETOS
ARG TARGETARCH
ADD --chmod=0555 https://github.com/argoproj/argo-cd/releases/download/v3.1.8/argocd-${TARGETOS}-${TARGETARCH} /usr/local/bin/argocd
COPY --from=builder /bin/diffbot /usr/local/bin/
