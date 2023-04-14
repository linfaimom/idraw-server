FROM golang:1.20-alpine as builder
LABEL MAINTAINER="Marcus Lin" MAIL="linfaimom@gmail.com"
WORKDIR /root/buildDir
COPY go.mod go.sum /root/buildDir/
RUN --mount=type=cache,target=/root/.cache/go-cache go mod download -x
COPY . /root/buildDir
RUN go build -v

FROM alpine as runner
LABEL MAINTAINER="Marcus Lin" MAIL="linfaimom@gmail.com"
WORKDIR /root/release
RUN --mount=type=cache,target=/root/.cache/apk-cache apk update && apk upgrade && apk add sqlite
COPY --from=builder  /root/buildDir/idraw-server /root/release/idraw-server
ENTRYPOINT ./idraw-server > application.log