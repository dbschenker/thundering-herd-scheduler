FROM --platform=$BUILDPLATFORM golang:1.17-alpine as builder

ARG RELEASE_VERSION=development

# Install our build tools
RUN apk add --update ca-certificates

WORKDIR /go/src/github.com/dbschenker/thundering-herd-scheduler
COPY . ./
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-X 'main.VERSION=${RELEASE_VERSION}'" -o bin/thundering-herd-scheduler ./cmd/thundering-herd-scheduler

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/dbschenker/thundering-herd-scheduler/bin/thundering-herd-scheduler /thundering-herd-scheduler

ENTRYPOINT ["/thundering-herd-scheduler"]
