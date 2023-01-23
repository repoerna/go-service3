SHELL := /bin/bash

run: 
	go run main.go

build:
	go build -ldflags "-X main.build=local"

# building containers
VERSION := 1.0

all: service

service:
	docker build \
		-f zarf/docker/dockerfile \
		-t service-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# k8s/kind
KIND_CLUSTER := repoerna-cluster

kind-up:
	kind create cluster\
		--image kindest/node:v1.26.0 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/config.yaml
	kubectl config set-context --current --namespace=service

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	kind load docker-image service-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	cat zarf/k8s/base/service-pod/base-service.yaml | kubectl apply -f -

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces
	# kubectl cluster-info --context kind-$(KIND_CLUSTER)

kind-status-service:
	kubectl get pods -o wide --watch --namespace=service-system

kind-logs:
	kubectl logs -l app=service --all-containers=true -f --tail=100 --namespace=service-system

kind-restart:
	kubectl rollout restart deployment service-pod --namespace=service-system
 
 kind-update: all kind-load kind-restart

 kind-describe:
	kubectl describe pod -l app=service --namespace=service-system