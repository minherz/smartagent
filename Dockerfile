# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.23rc1-alpine as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY *.go ./
COPY metadata/ metadata/
COPY utils/ utils/

# Build the binary.
RUN CGO_ENABLED=0 go build -mod=readonly -installsuffix 'static' -v -o smartagent

# Use empty image for a lean production container.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM scratch

# Copy certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/smartagent /app/smartagent
COPY static/ /app/static/

# Run the web service on container startup.
CMD ["/app/smartagent"]