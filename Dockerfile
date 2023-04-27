FROM golang:1.20.3-alpine as builder

USER root

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

RUN adduser -D -g '' appuser

WORKDIR $GOPATH/src/mypackage/myapp/

COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/echoserver

############################
# STEP 2 build a small image
############################
FROM scratch

COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable.
COPY --from=builder /go/bin/echoserver /go/bin/echoserver

# Use an unprivileged user.
USER appuser

HEALTHCHECK --interval=5s --timeout=3s --retries=10 --start-period=5s \
    CMD ["/go/bin/echoserver", "healthcheck"]

# Port on which the service will be exposed.
EXPOSE 8080

# Run the hello binary.
ENTRYPOINT ["/go/bin/echoserver"]
