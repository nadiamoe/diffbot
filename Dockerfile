FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:f1ddd9fe14fffc091dd98cb4bfa999f32c5fc77d2f2305ea9f0e2595c5437c14 as builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /diffbot

COPY . .
# go env GOCACHE; go env GOMODCACHE
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/diffbot .

FROM alpine:3.24.0@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4

ARG TARGETOS
ARG TARGETARCH
ADD --chmod=0555 https://github.com/argoproj/argo-cd/releases/download/v3.4.3/argocd-${TARGETOS}-${TARGETARCH} /usr/local/bin/argocd
COPY --from=builder /bin/diffbot /usr/local/bin/
