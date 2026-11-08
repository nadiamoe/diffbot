FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 as builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /diffbot

COPY . .
# go env GOCACHE; go env GOMODCACHE
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/diffbot .

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

ARG TARGETOS
ARG TARGETARCH
ADD --chmod=0555 https://github.com/argoproj/argo-cd/releases/download/v3.6.1/argocd-${TARGETOS}-${TARGETARCH} /usr/local/bin/argocd
COPY --from=builder /bin/diffbot /usr/local/bin/
