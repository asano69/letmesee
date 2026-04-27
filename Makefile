.PHONY: build
build:
	go build -o letmesee .

.PHONY: server
server:
	./letmesee -config config.yaml -listen :8080

.PHONY: build-image
build-image: ## Build Docker image
	docker build -t registry.internal/go-letmesee:latest .


.PHONY: push-image
push-image: ## Push Docker image
	docker push registry.internal/go-letmesee:latest

.PHONY: deploy
deploy: build-image push-image ## (*) Deploy stack via Komodo
	docker exec -it komodo km x -y destroy-stack letmesee
	docker exec -it komodo km x -y pull-stack   letmesee
	docker exec -it komodo km x -y deploy-stack letmesee
