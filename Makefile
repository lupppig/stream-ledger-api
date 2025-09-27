
# Postgres
POSTGRES_IMAGE=postgres:15
POSTGRES_CONTAINER=postgres
POSTGRES_PORT=5432
POSTGRES_USER=user
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=mydb

# Redis
REDIS_IMAGE=redis:7
REDIS_CONTAINER=redis
REDIS_PORT=6379

# Kafka (using Bitnami image for simplicity)
KAFKA_IMAGE=bitnami/kafka:3.7.0
KAFKA_CONTAINER=kafka
KAFKA_PORT=9092
KAFKA_ZK_PORT=2181

# Silent by default
.SILENT:

# ==============================
# Postgres
# ==============================
postgres-pull:
	docker pull $(POSTGRES_IMAGE)

postgres-up:
	docker run -d --rm \
		--name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		$(POSTGRES_IMAGE)
psql:
	docker exec -it postgres psql -U user -d $(POSTGRES_DB)
postgres-down:
	docker stop $(POSTGRES_CONTAINER) || true

postgres-test:
	docker run --rm -d \
		--name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		postgres:15

# Stop the test Postgres container
postgres-test-down:
	docker stop $(POSTGRES_CONTAINER)

postgres-test-url:
	@echo "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

# ==============================
# Kafka
# ==============================
kafka-pull:
	docker pull $(KAFKA_IMAGE)

kafka-up:
	docker run -d --rm \
		--name $(KAFKA_CONTAINER) \
		-e KAFKA_CFG_LISTENERS=PLAINTEXT://:$(KAFKA_PORT) 
		-e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:$(KAFKA_PORT) \
		-e ALLOW_PLAINTEXT_LISTENER=yes \
		-p $(KAFKA_PORT):$(KAFKA_PORT) \
		$(KAFKA_IMAGE)

kafka-down:
	docker stop $(KAFKA_CONTAINER) || true

# ==============================
# Convenience
# ==============================
up: postgres-up kafka-up
down: postgres-down  kafka-down
pull: postgres-pull kafka-pull
restart: down up

run:
	go run .

test: postgres-test
	go test ./tests/... -v -count=1