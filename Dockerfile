FROM golang:latest as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN go get github.com/PuerkitoBio/goquery \
    github.com/alyu/configparser \
    github.com/mattn/go-sqlite3 \
    github.com/sevlyar/go-daemon && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o neagent . && \
    useradd -M -N -o -u 1000 neagent
FROM scratch
COPY --from=builder /build/neagent /app/
COPY --from=builder /etc/passwd /etc/passwd
WORKDIR /app
USER neagent
ENTRYPOINT ["/app/neagent"]
