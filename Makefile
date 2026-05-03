.PHONY: build lb backend compose up down test lint fmt

build: lb/backend image-backend image-lb

image-lb:
	docker build -t predictive-sentinel-lb:local ./lb

image-backend:
	docker build -t predictive-sentinel-backend:local ./backend

up:
	docker-compose up --build -d

down:
	docker-compose down

logs:
	docker-compose logs -f

test:
	go test ./...

fmt:
	gofmt -w ./lb

lint:
	# placeholder for golangci-lint
	echo "Run golangci-lint locally"
