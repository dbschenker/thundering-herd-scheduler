all: build test

version := v1.30.0-0

build:
	go build -o bin/thundering-herd-scheduler ./cmd/thundering-herd-scheduler/main.go

build-version:
		go build \
		 	-o bin/thundering-herd-scheduler \
		  -ldflags="-X 'main.Version=$(version)' \
			-X 'k8s.io/component-base/version.gitVersion=$(version)'" \
			./cmd/thundering-herd-scheduler/main.go

build-debug:
		go build -gcflags="all=-N -l" -o bin/thundering-herd-scheduler-debug ./cmd/thundering-herd-scheduler/main.go
# To run debug: dlv --listen=:2345 --headless=true --api-version=2 exec bin/thundering-herd-scheduler-debug --ARGS

clean:
		rm -r bin/

test:
		go test ./...

docker:
		docker buildx build -t thundering-herd-scheduler:local -t thundering-herd-scheduler:${version} --load --build-arg RELEASE_VERSION=${version} .

local:
	bin/thundering-herd-scheduler --config manifests/development/scheduler.yaml

kind:
	kind create cluster --config manifests/development/kind-config.yaml
	kind get kubeconfig > deployment/config
e2e-test:
	kind load docker-image thundering-herd-scheduler:${version}
	IMAGE_TAG=${version} chainsaw test
