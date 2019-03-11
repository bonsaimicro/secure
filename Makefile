GO := go
GLIDE := glide
COMPOSE := docker-compose
TAG?=$(shell git rev-list HEAD --max-count=1 --abbrev-commit)
export TAG

test:
	@$(GO) test ./...

setup:
	@$(GLIDE) install

run:
	@DB_PATH=/tmp/badger \
	PORT=1234 \
	ENV=development \
	DARE_PASSWORD=password \
	DARE_SALT=salt \
	JWT_SECRET=jwt_secret \
	$(GO) run main.go

run-docker:
	@$(COMPOSE) up --build

compile:
	@docker build -f Dockerfile.prd -t tmp . && docker run -t -i --mount type=bind,src=`pwd`,dst=/binary tmp
