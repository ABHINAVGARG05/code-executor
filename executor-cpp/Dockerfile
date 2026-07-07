FROM golang:1.25-alpine AS build
WORKDIR /app
COPY shared ./shared
COPY executor-cpp/go.mod executor-cpp/go.sum ./executor-cpp/
WORKDIR /app/executor-cpp
RUN go mod download
COPY executor-cpp/. .
RUN CGO_ENABLED=0 go build -o /bin/executor-cpp .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /workspace
COPY --from=build /bin/executor-cpp /usr/local/bin/executor-cpp
ENTRYPOINT ["executor-cpp"]
