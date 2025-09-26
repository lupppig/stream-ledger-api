
# Postgres
POSTGRES_IMAGE=postgres:15
POSTGRES_CONTAINER=my_postgres
POSTGRES_PORT=5432
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=mydb

# Redis
REDIS_IMAGE=redis:7
REDIS_CONTAINER=my_redis
REDIS_PORT=6379

# Kafka (using Bitnami image for simplicity)
KAFKA_IMAGE=bitnami/kafka:3.7.0
KAFKA_CONTAINER=my_kafka
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

postgres-down:
	docker stop $(POSTGRES_CONTAINER) || true

# ==============================
# Redis
# ==============================
redis-pull:
	docker pull $(REDIS_IMAGE)

redis-up:
	docker run -d --rm \
		--name $(REDIS_CONTAINER) \
		-p $(REDIS_PORT):6379 \
		$(REDIS_IMAGE)

redis-down:
	docker stop $(REDIS_CONTAINER) || true

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
up: postgres-up redis-up kafka-up
down: postgres-down redis-down kafka-down
pull: postgres-pull redis-pull kafka-pull
restart: down up

run: go run .