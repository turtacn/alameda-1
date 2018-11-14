# Build the manager binary
FROM golang:1.11.2 as builder

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/alameda/operator
Add . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/containers-ai/alameda/operator/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/containers-ai/alameda/operator/manager .
EXPOSE 50050/tcp
ENTRYPOINT ["./manager"]
