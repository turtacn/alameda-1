# Build the manager binary
FROM golang:1.11.4-alpine as builder

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/alameda
ADD . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ./datahub/datahub github.com/containers-ai/alameda/datahub/cmd

# Copy the controller-manager into a thin image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/etc/datahub.yml /etc/alameda/datahub/datahub.yml
COPY --from=builder /go/src/github.com/containers-ai/alameda/datahub/datahub .
EXPOSE 50050/tcp
ENTRYPOINT ["./datahub"]
CMD [ "run" ]
