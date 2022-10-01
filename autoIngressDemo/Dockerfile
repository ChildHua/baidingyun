# hdrun911/ingress-manager:1.0.0
FROM golang:1.18.6 AS builder
WORKDIR /app
COPY . .
# RUN go env 环境变量设置
RUN CGO_ENABLED=0 go build -o ingress-manager autoIngressDemo/main.go

FROM alpine:3.15.3

WORKDIR /app

COPY --from=builder /app/ingress-manager .

CMD ["./ingress-manager"]