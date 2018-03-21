FROM golang as builder
RUN go get github.com/tianon/gosu && \
    cd /go/src/github.com/tianon/gosu && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /root/gosu .
WORKDIR /go/src/github.com/kiyor/playlist
COPY *.go ./
RUN go get && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /root/playlist .

FROM centos:7
ENV LANG en_US.UTF-8
RUN rpm -Uvh http://repository.it4i.cz/mirrors/repoforge/redhat/el7/en/x86_64/rpmforge/RPMS/rpmforge-release-0.5.3-1.el7.rf.x86_64.rpm && \
    rpm -Uvh http://li.nux.ro/download/nux/dextop/el7/x86_64/nux-dextop-release-0-1.el7.nux.noarch.rpm && \
    yum install -y epel-release && \
    yum install -y ffmpeg perl file
WORKDIR /root
COPY --from=builder /root/gosu .
COPY --from=builder /root/playlist .
COPY run.sh .
COPY bin/ass2srt.pl /usr/local/bin/

ENTRYPOINT ["/root/run.sh"]
