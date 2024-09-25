# build environment
FROM golang:1.22 AS build-env
WORKDIR /server
COPY src/go.mod ./
RUN go mod download
COPY src src
WORKDIR /server/src
RUN CGO_ENABLED=0 GOOS=linux go build -o /server/build/httpserver .

FROM alpine:3.15
WORKDIR /app
RUN mkdir tmp
#RUN set -x \
#    && apk add --no-cache ca-certificates tzdata \
#    && cp /usr/share/zoneinfo/Europe/Kiev /etc/localtime \
#    && echo Europe/Kiev > /etc/timezone \
#    && apk del tzdata

COPY --from=build-env /server/build/httpserver /app/dle-proxy

#ENV GITHUB-SHA=<GITHUB-SHA>

#EXPOSE 28091/tcp

ENTRYPOINT [ "/app/dle-proxy" ]
#ENTRYPOINT [ "ls", "-la", "/app/httpserver" ]
