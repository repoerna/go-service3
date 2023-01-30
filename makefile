SHELL := /bin/bash

run: 
	go run cmd/services/sales-api/main.go | go run cmd/tooling/logfmt/main.go 

build:
	CGO_ENABLED=0 go build -ldflags "-X main.build=local"

# ==============================================================================
# building containers
VERSION := 1.0

all: sales-api

sales-api:
	docker build \
		-f deployment/docker/dockerfile.sales-api \
		-t sales-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# k8s/kind
KIND_CLUSTER := repoerna-cluster

kind-up:
	kind create cluster\
		--image kindest/node:v1.26.0 \
		--name $(KIND_CLUSTER) \
		--config deployment/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	cd deployment/k8s/kind/sales-pod; kustomize edit set image sales-api-image=sales-api-amd64:$(VERSION)
	kind load docker-image sales-api-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	# cat deployment/k8s/base/sales-pod/base-sales.yaml | kubectl apply -f -
	kustomize build deployment/k8s/kind/sales-pod | kubectl apply -f -

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces
	# kubectl cluster-info --context kind-$(KIND_CLUSTER)

kind-status-sales:
	kubectl get pods -o wide --watch --namespace=sales-system

kind-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 --namespace=sales-system | go run cmd/tooling/logfmt/main.go

kind-restart:
	kubectl rollout restart deployment sales-pod --namespace=sales-system
 
 kind-update: all kind-load kind-restart

 kind-update-apply: all kind-load kind-apply

 kind-describe:
	kubectl describe pod -l app=sales --namespace=sales-system