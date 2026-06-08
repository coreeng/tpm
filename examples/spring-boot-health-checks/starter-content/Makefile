IMAGE_NAME ?= health-app
IMAGE_TAG ?= local
IMAGE ?= $(IMAGE_NAME):$(IMAGE_TAG)
KIND_CLUSTER ?= kind
REGISTRY ?=
WORKSPACE_NAMESPACE ?=

.PHONY: build kind-load push deploy
# Build the Spring Boot app into a container image with a multi-stage Docker build
# (Gradle/JDK21 builder -> JRE runtime). Produces $(IMAGE) in the local Docker daemon.
build:
	docker build -t "$(IMAGE)" .

kind-load: build
	kind load docker-image "$(IMAGE)" --name "$(KIND_CLUSTER)"

push: build
	@test -n "$(REGISTRY)" || (echo "REGISTRY is required" >&2 && exit 1)
	docker tag "$(IMAGE)" "$(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"
	docker push "$(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)"

deploy:
	@test -n "$(WORKSPACE_NAMESPACE)" || (echo "WORKSPACE_NAMESPACE is required" >&2 && exit 1)
	sed "s|image: health-app:local|image: $(IMAGE)|" k8s/deployment.yaml | kubectl apply -n "$(WORKSPACE_NAMESPACE)" -f -
	kubectl apply -n "$(WORKSPACE_NAMESPACE)" -f k8s/service.yaml
