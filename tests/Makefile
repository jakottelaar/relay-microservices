COMPOSE_TEST = docker compose -f ./infrastructure/docker-compose.test.yml

.PHONY: test-up test-down test-logs test test-run

test-up:
	$(COMPOSE_TEST) up -d --build

test-down:
	$(COMPOSE_TEST) down -v

test-logs:
	$(COMPOSE_TEST) logs -f

test: test-up
	@python -m pytest e2e/ -v & $(MAKE) test-down

test-run:
	python -m pytest e2e/ -v