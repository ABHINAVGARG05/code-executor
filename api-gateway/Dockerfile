# API Gateway Dockerfile
FROM golang:1.25-alpine AS build
WORKDIR /app
# Copy shared module and api-gateway go mod first for caching
COPY shared ./shared
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /app/api-gateway
RUN go mod download
# Copy source
COPY api-gateway/. .
RUN go build -o /bin/api-gateway .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /bin/api-gateway /usr/local/bin/api-gateway
EXPOSE 8080
ENTRYPOINT ["api-gateway"]
