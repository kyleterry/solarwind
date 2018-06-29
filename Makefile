RELEASE ?= kyleterry/solarwind:latest
DOCKER  ?= docker

build:
	$(DOCKER) build -t $(RELEASE) .

release:
	@echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	$(DOCKER) push $(RELEASE)
