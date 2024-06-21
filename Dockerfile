FROM --platform=$BUILDPLATFORM golang:1.21.11-alpine as builder

ARG RELEASE_VERSION=development

# Install our build tools
RUN apk add --update ca-certificates

WORKDIR /go/src/github.com/dbschenker/thundering-herd-scheduler

ARG TARGETOS
ARG TARGETARCH
ENV LDFLAGS "-X 'main.VERSION=${RELEASE_VERSION}' "

COPY . ./

RUN if echo "$RELEASE_VERSION" | grep -Eq '^v\d+\.\d+\.\d+.*'; then export LDFLAGS="$LDFLAGS -X 'k8s.io/component-base/version.gitVersion=${RELEASE_VERSION}'"; fi  && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="$LDFLAGS" -o bin/thundering-herd-scheduler ./cmd/thundering-herd-scheduler

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/dbschenker/thundering-herd-scheduler/bin/thundering-herd-scheduler /thundering-herd-scheduler

ENTRYPOINT ["/thundering-herd-scheduler"]
