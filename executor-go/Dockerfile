# Go executor Dockerfile
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY shared ./shared
COPY executor-go/go.mod executor-go/go.sum ./executor-go/
WORKDIR /app/executor-go
RUN go mod download
COPY executor-go/. .
RUN go build -o /bin/executor-go .

FROM golang:1.25-alpine
# Need Go toolchain to run user code via `go run`
WORKDIR /workspace
COPY --from=build /bin/executor-go /usr/local/bin/executor-go
ENTRYPOINT ["executor-go"]
