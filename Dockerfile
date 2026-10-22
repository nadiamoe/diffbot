FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 as builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /diffbot

COPY . .
# go env GOCACHE; go env GOMODCACHE
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/diffbot .

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

ARG TARGETOS
ARG TARGETARCH
ADD --chmod=0555 https://github.com/argoproj/argo-cd/releases/download/v3.5.6/argocd-${TARGETOS}-${TARGETARCH} /usr/local/bin/argocd
COPY --from=builder /bin/diffbot /usr/local/bin/
