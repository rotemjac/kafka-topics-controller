# import deploy config
# You can change the default deploy config with `make cnf="deploy_special.env" release`
dpl ?= deploy.env
include $(dpl)
export $(shell sed 's/=.*//' $(dpl))

dpl_root ?= ../deploy.env
include $(dpl_root)
export $(shell sed 's/=.*//' $(dpl_root))

include ../Makefile


build:
	go mod tidy
	go mod vendor
	cd cmd && CGO_ENABLED=0 go build -o ../artifacts/${BINARY_NAME}

# Build the container
build-tag: ## Build the release and development container.
	docker build -t $(DOCKER_REPO_NAME) .
	docker tag $(DOCKER_REPO_NAME):latest $(DOCKER_REPO_PATH)/$(DOCKER_REPO_NAME):$(DOCKER_TAG)

push: build-tag
	docker push $(DOCKER_REPO_PATH)/$(DOCKER_REPO_NAME):$(DOCKER_TAG)


#for official branches
build-fetcher:
	go mod tidy
	go mod vendor
	cd cmd && CGO_ENABLED=0 go build -o ../artifacts/${BINARY_NAME}

image: build-fetcher
	docker build -f build/Dockerfile -t ${DOCKER_REPO_PATH}/${DOCKER_REPO_NAME}:latest .
	docker tag ${DOCKER_REPO_PATH}/${DOCKER_REPO_NAME}:latest ${DOCKER_REPO_PATH}/${DOCKER_REPO_NAME}:${DOCKER_TAG}
#	docker tag ${DOCKER_REPO_PATH}/${DOCKER_REPO_NAME}:latest ${DOCKER_REPO_PATH}/${DOCKER_REPO_NAME}:${BINARY_NAME}-${BUILD_DATE}

public-image: auth
	docker push ${DOCKER_REPO_PATH}/${DOCKER_REPO_NAME}:latest
