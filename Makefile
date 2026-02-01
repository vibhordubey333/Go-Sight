.PHONY: run compose-up compose-down compose-clean k8s-apply k8s-delete k8s-status k8s-port-forward k8s-build helm-install helm-uninstall

run:
	go run .

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down --remove-orphans

compose-clean:
	docker compose down --remove-orphans --volumes

k8s-build:
	eval $$(minikube -p minikube docker-env) && docker build -t go-api:latest .

helm-install:
	helm install go-sight ./helm/go-sight --namespace go-sight-backend --create-namespace

helm-uninstall:
	helm uninstall go-sight --namespace go-sight-backend
k8s-apply:
	kubectl apply -f k8s/00-namespace.yaml
	kubectl apply -f k8s

k8s-delete:
	kubectl delete -f k8s

k8s-status:
	kubectl get -n go-sight-backend pods,svc

k8s-port-forward:
	kubectl port-forward svc/api 8000:8000