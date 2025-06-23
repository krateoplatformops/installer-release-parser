ARCH?=amd64
REPO?=#your repository here 
VERSION?=0.1

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o ./bin/installer-release-parser main.go

container:
	docker build -t $(REPO)installer-release-parser:$(VERSION) .
	docker push $(REPO)installer-release-parser:$(VERSION)

container-multi:
	docker buildx build --tag $(REPO)installer-release-parser:$(VERSION) --push --platform linux/amd64,linux/arm64 .
