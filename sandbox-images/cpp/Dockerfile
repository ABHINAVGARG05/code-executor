FROM alpine:3.20

RUN apk add --no-cache g++ && adduser -D -u 1000 sandbox

USER sandbox
WORKDIR /home/sandbox
