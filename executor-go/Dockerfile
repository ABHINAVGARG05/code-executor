FROM golang:1.25-alpine AS build
WORKDIR /app
COPY shared ./shared
COPY executor-go/go.mod executor-go/go.sum ./executor-go/
WORKDIR /app/executor-go
RUN go mod download
COPY executor-go/. .
RUN CGO_ENABLED=0 go build -o /bin/executor-go .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /workspace
COPY --from=build /bin/executor-go /usr/local/bin/executor-go
ENTRYPOINT ["executor-go"]
