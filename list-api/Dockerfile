FROM golang:alpine AS build

ARG app_version
ARG commit_sha

ENV VERSION=$app_version
ENV COMMIT_SHA=$commit_sha

RUN apk update --no-cache && apk add git
WORKDIR /app

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

ADD ./ /app

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Ver=${VERSION} -X main.SHA=${COMMIT_SHA} -w -extldflags '-static'" -o golang-api

FROM alpine
WORKDIR /app
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/
COPY --from=build /app/golang-api /app

ENTRYPOINT ["/app/golang-api"]

EXPOSE 8080
