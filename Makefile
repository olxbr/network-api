.PHONY: test build

GOLANGLINT_INSTALLED_VERSION := $(shell golangci-lint version 2>/dev/null | sed -ne 's/.*version\ \([0-9]*\.[0-9]*\.[0-9]*\).*/\1/p')
GOLANG_LINT_VERSION := 2.1.6

REPO := github.com/olxbr/network-api

COMMIT_ID := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_TAG := $(shell git describe --tags --always --abbrev=0 --match="v[0-9]*.[0-9]*.[0-9]*" 2> /dev/null)
VERSION := $(shell echo "${COMMIT_TAG}" | sed 's/^.//')

GO_LDFLAGS := -ldflags "-X '${REPO}/cmd.Version=${VERSION}' -X '${REPO}/cmd.CommitID=${COMMIT_ID}' -X '${REPO}/cmd.BuildTime=${BUILD_TIME}'"

test:
	@go test ./... -coverprofile cover.out
	@echo "" && go tool cover -func cover.out | grep -e '^total.*' | tr -s '[:blank:]' ' '
	@rm cover.out

lint:
ifneq (${GOLANG_LINT_VERSION}, ${GOLANGLINT_INSTALLED_VERSION})
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v${GOLANG_LINT_VERSION}
endif
	$$(go env GOPATH)/bin/golangci-lint run

build:
	go build -o ./bin/network-api ${GO_LDFLAGS} ./cmd/network-api

build_cli:
	go build -o ./bin/network-cli ${GO_LDFLAGS} ./cmd/network-cli

install_cli:
	go install ${GO_LDFLAGS} ./cmd/network-cli

mocks:
	cd pkg/db && mockery --all --output fake --outpkg fake --case underscore
	cd pkg/secret && mockery --all --output fake --outpkg fake --case underscore

run:
	GOARCH=amd64 GOOS=linux go build -o deployment/network-api ${GO_LDFLAGS} ./cmd/network-api
	sam local start-api \
	--template-file deployment/sam_network_api.yaml \
	--docker-network network-api_default

clean:
	rm -rf ./bin/*
	rm -rf deployment/aws-provider
	rm -rf deployment/network-api
	rm -rf deployment/jwt-authorizer
	mkdir -p deployment/network-api
	mkdir -p deployment/jwt-authorizer
	mkdir -p deployment/aws-provider

package: clean
	GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o deployment/network-api/bootstrap ${GO_LDFLAGS} ./cmd/network-api
	GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o deployment/jwt-authorizer/bootstrap ${GO_LDFLAGS} ./cmd/jwt-authorizer
	sam package --template-file deployment/sam_network_api.yaml --s3-bucket network-api-sam --output-template-file packaged.yaml

deploy:
	set -e ; \
	SAM_PARAMETERS=$$(cat parameters.json | jq -r '[ .[] | "\(.ParameterKey)=\(.ParameterValue)" ] | join(" ")' ) ; \
	sam deploy --template-file packaged.yaml --stack-name network-api-sam \
	--parameter-overrides $$SAM_PARAMETERS \
	--capabilities CAPABILITY_IAM

package_provider: clean
	GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o deployment/aws-provider/bootstrap ${GO_LDFLAGS} ./cmd/aws-provider
	sam package --template-file deployment/sam_aws_provider.yaml --s3-bucket network-api-sam --output-template-file packaged-provider.yaml

deploy_provider:
	set -e ; \
	SAM_PARAMETERS=$$(cat parameters_provider.json | jq -r '[ .[] | "\(.ParameterKey)=\(.ParameterValue)" ] | join(" ")' ) ; \
	sam deploy --template-file packaged-provider.yaml --stack-name network-provider-sam \
	--parameter-overrides $$SAM_PARAMETERS \
	--capabilities CAPABILITY_IAM
