# C++ executor Dockerfile
FROM golang:1.25-alpine AS gobuild
WORKDIR /app
COPY shared ./shared
COPY executor-cpp/go.mod executor-cpp/go.sum ./executor-cpp/
WORKDIR /app/executor-cpp
RUN go mod download
COPY executor-cpp/. .
RUN go build -o /bin/executor-cpp .

FROM alpine:3.20
RUN apk add --no-cache g++
WORKDIR /workspace
COPY --from=gobuild /bin/executor-cpp /usr/local/bin/executor-cpp
ENTRYPOINT ["executor-cpp"]
