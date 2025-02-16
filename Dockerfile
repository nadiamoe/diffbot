FROM golang:1.24-alpine@sha256:2d40d4fc278dad38be0777d5e2a88a2c6dee51b0b29c97a764fc6c6a11ca893c as builder

WORKDIR /diffbot

RUN wget -O argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/download/v2.14.2/argocd-linux-amd64 \
    && install -m 555 argocd-linux-amd64 /bin/argocd

COPY . .
RUN go build -o /bin/diffbot .

FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c

COPY --from=builder /bin/diffbot /usr/local/bin/
COPY --from=builder /bin/argocd /usr/local/bin/
