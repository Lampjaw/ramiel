FROM golang:1.18-alpine AS build-env

RUN apk add -U --no-cache build-base git

RUN mkdir /app
RUN mkdir /build
WORKDIR /app

ADD ./src /app

RUN go get -d ./... && \
    go build -v -o /build ./cmd/ramiel

FROM alpine:latest

RUN apk add -U --no-cache iputils ca-certificates tzdata

COPY --from=build-env /build /bin

CMD [ "/bin/ramiel" ]