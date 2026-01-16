.PHONY: run compose-up compose-down compose-clean k8s-apply k8s-delete k8s-status k8s-port-forward k8s-build

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
k8s-apply:
	kubectl apply -f k8s/00-namespace.yaml
	kubectl apply -f k8s

k8s-delete:
	kubectl delete -f k8s

k8s-status:
	kubectl get pods,svc

k8s-port-forward:
	kubectl port-forward svc/api 8000:8000