# Postgres
POSTGRES_IMAGE=postgres:15
POSTGRES_CONTAINER=postgres
POSTGRES_PORT=5432
POSTGRES_USER=user
POSTGRES_TEST_CONTAINER=postgres_test
POSTGRES_TEST_PORT=5044
POSTGRES_TEST_DB=mydbtest
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=mydb


# Kafka (using Bitnami image for simplicity)
KAFKA_IMAGE=confluentinc/cp-kafka:7.5.0 
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
	@echo "Starting PostgreSQL test instance on port $(POSTGRES_TEST_PORT)..."
	docker run -d \
		--name $(POSTGRES_TEST_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_TEST_DB) \
		-p $(POSTGRES_TEST_PORT):5432 \
		postgres:15
	@echo "Waiting for PostgreSQL to be ready..."
	@until nc -z 127.0.0.1 $(POSTGRES_TEST_PORT); do sleep 1; done
	@echo "PostgreSQL ready"


psql-test:
	docker exec -it postgres_test psql -U user -d $(POSTGRES_TEST_DB)
# Stop the test Postgres container
postgres-test-down:
	docker stop $(POSTGRES_TEST_CONTAINER)

postgres-test-url:
	@echo "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"

# ==============================
# Kafka
# ==============================
kafka-pull:
	docker pull $(KAFKA_IMAGE)

ZOOKEEPER_CONTAINER=zookeeper
ZOOKEEPER_IMAGE=confluentinc/cp-zookeeper:7.5.0
NETWORK_NAME = kafka-network


kafka-up:
	# Create Docker network
	docker network create $(NETWORK_NAME) || true
	
	# Start Zookeeper
	docker run -d --rm \
		--name $(ZOOKEEPER_CONTAINER) \
		--network $(NETWORK_NAME) \
		-e ZOOKEEPER_CLIENT_PORT=2181 \
		-e ZOOKEEPER_TICK_TIME=2000 \
		-p $(ZOOKEEPER_PORT):2181 \
		$(ZOOKEEPER_IMAGE)
	
	# Wait for Zookeeper to be ready
	sleep 10
	
	# Start Kafka
	docker run -d --rm \
		--name $(KAFKA_CONTAINER) \
		--network $(NETWORK_NAME) \
		-e KAFKA_BROKER_ID=1 \
		-e KAFKA_ZOOKEEPER_CONNECT=$(ZOOKEEPER_CONTAINER):2181 \
		-e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:$(KAFKA_PORT) \
		-e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:$(KAFKA_PORT) \
		-e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT \
		-e KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT \
		-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
		-e KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
		-p $(KAFKA_PORT):$(KAFKA_PORT) \
		$(KAFKA_IMAGE)
	
	# Wait for Kafka to be ready
	sleep 15
	@echo "Kafka and Zookeeper are starting up..."

kafka-down:
	docker stop $(KAFKA_CONTAINER) $(ZOOKEEPER_CONTAINER) || true
	docker network rm $(NETWORK_NAME) || true

kafka-status:
	@echo "=== Container Status ==="
	docker ps | grep -E "(kafka|zookeeper)" || echo "No Kafka/Zookeeper containers running"
	@echo "=== Network Status ==="
	docker network ls | grep $(NETWORK_NAME) || echo "Network not found"

# ==============================
# Convenience
# ==============================
up: postgres-up kafka-up
down: postgres-down  kafka-down
pull: postgres-pull kafka-pull
restart: down up

run:
	go run .

test:
	go test ./tests/... -v -count=1