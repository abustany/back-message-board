# golang 1.12.7-buster
FROM golang@sha256:55803225abf9cdc5b42c913d5d8c8f2add70ae101650d64a5f92fdf685309b5a AS build
#WORKDIR /go/src/github.com/abustany/back-message-board
WORKDIR /build
COPY . .
#RUN GOFLAGS=-mod=vendor CGO_ENABLED=0 GOOS=linux make adminserver
RUN CGO_ENABLED=0 GOOS=linux make

# alpine 3.10.1
FROM alpine@sha256:6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
RUN adduser -h /home/server -D server
WORKDIR /home/server
COPY --from=build /build/server .
COPY docker-entrypoint.sh .
USER server
EXPOSE 1412
CMD ["/home/server/docker-entrypoint.sh"]
