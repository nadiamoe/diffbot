FROM golang:1.21-alpine as builder

WORKDIR /diffbot

COPY . .
RUN go build -o /bin/diffbot .

FROM alpine:3.19.1

RUN apk add git curl
RUN curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.9.6/argocd-linux-amd64 \
    && install -m 555 argocd-linux-amd64 /usr/local/bin/argocd \
    && rm argocd-linux-amd64

COPY --from=builder /bin/diffbot /usr/local/bin/
