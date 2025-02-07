FROM golang:1.23-alpine@sha256:2c49857f2295e89b23b28386e57e018a86620a8fede5003900f2d138ba9c4037 as builder

WORKDIR /diffbot

RUN wget -O argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.14.2/argocd-linux-amd64 \
    && install -m 555 argocd-linux-amd64 /bin/argocd

COPY . .
RUN go build -o /bin/diffbot .

FROM alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099

COPY --from=builder /bin/diffbot /usr/local/bin/
COPY --from=builder /bin/argocd /usr/local/bin/
