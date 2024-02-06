FROM golang:1.21-alpine as builder

WORKDIR /diffbot

RUN wget -O argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.10.0/argocd-linux-amd64 \
    && install -m 555 argocd-linux-amd64 /bin/argocd

COPY . .
RUN go build -o /bin/diffbot .

FROM alpine:3.19.1

COPY --from=builder /bin/diffbot /usr/local/bin/
COPY --from=builder /bin/argocd /usr/local/bin/
