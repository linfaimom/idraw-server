FROM golang:1.20-alpine as builder
LABEL MAINTAINER="Marcus Lin" MAIL="linfaimom@gmail.com"
WORKDIR /root/buildDir
COPY ./ /root/buildDir
RUN go env -w GOPROXY=https://goproxy.cn,direct && go build -v

FROM alpine as runner
LABEL MAINTAINER="Marcus Lin" MAIL="linfaimom@gmail.com"
WORKDIR /root/release
COPY --from=builder  /root/buildDir/idraw-server /root/release/idraw-server
RUN mkdir -pv /data/logs
ENTRYPOINT ./idraw-server > /data/logs/application.log