# Build the manager binary
FROM golang:1.11.5-stretch as builder

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/alameda
ADD . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./bin/admission-controller github.com/containers-ai/alameda/admission-controller

# Copy the controller-manager into a thin image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/containers-ai/alameda/bin/admission-controller .
COPY --from=builder /go/src/github.com/containers-ai/alameda/admission-controller/etc/admission-controller.yml /etc/alameda/admission-controller/admission-controller.yml
EXPOSE 443/tcp
ENTRYPOINT ["./admission-controller"]
