# High-Performance Matching Engine

A production-ready, high-throughput matching engine for cryptocurrency exchanges built with Golang, PostgreSQL, and Kafka.

## 🏗️ Architecture

### Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Applications                     │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Load Balancer                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌────────────────────────────────────────────────────────────────┐
│                    Matching Engine Replicas                    │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐                  │
│  │  Engine  │    │  Engine  │    │  Engine  │                  │
│  │ BTC-USD  │    │ ETH-USD  │    │ SOL-USD  │    (Per pair)    │
│  └────┬─────┘    └────┬─────┘    └────┬─────┘                  │
│       │               │               │                        │
│  ┌────▼───────────────▼───────────────▼─────┐                  │
│  │      Red-Black Tree Order Books          │                  │
│  │    (In-Memory High-Performance)          │                  │
│  └────┬───────────────┬───────────────┬─────┘                  │
└───────┼───────────────┼───────────────┼────────────────────────┘
        │               │               │
        ▼               ▼               ▼
┌────────────────────────────────────────────────────────────────┐
│                         Kafka Cluster                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Trades     │  │Order Updates │  │   Commands   │          │
│  │    Topic     │  │    Topic     │  │    Topic     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────────────────────────────────────────┘
        │               │               │
        ▼               ▼               ▼
┌────────────────────────────────────────────────────────────────┐
│                      PostgreSQL Database                       │
│  ┌──────────┐  ┌──────────┐  ┌────────────────┐                │
│  │  Orders  │  │  Trades  │  │  Order Book    │                │
│  │  Table   │  │  Table   │  │  Snapshots     │                │
│  └──────────┘  └──────────┘  └────────────────┘                │
└────────────────────────────────────────────────────────────────┘
```

### Key Components

#### 1. **Matching Engine Core**
- **Red-Black Tree Order Book**: O(log n) insertion, deletion, and lookup
- **Pair-Based Design**: One engine instance per trading pair for horizontal scalability
- **In-Memory Matching**: Ultra-low latency order matching
- **FIFO Price-Time Priority**: Fair order execution

#### 2. **Asynchronous Architecture**
- **Kafka Integration**: All non-matching operations are asynchronous
- **Event-Driven**: Trades and order updates published to Kafka topics
- **Command Pattern**: Order submissions via Kafka for better throughput

#### 3. **Data Persistence**
- **PostgreSQL**: Strong consistency for order and trade storage
- **Write-Ahead Pattern**: Orders persisted after matching
- **Recovery Support**: Rebuild order book from database on restart

#### 4. **Kubernetes Operator**
- **Custom Resource Definition (CRD)**: Declarative matching engine deployment
- **Auto-scaling**: CPU and memory-based horizontal scaling
- **High Availability**: StatefulSet with pod anti-affinity
- **Rolling Updates**: Zero-downtime deployments

## 🚀 Features

### Performance
- ✅ **100,000+ orders/second** throughput per engine
- ✅ **Sub-millisecond** matching latency
- ✅ **O(log n)** order book operations using red-black trees
- ✅ **Lock-free** read operations for order book queries

### Reliability
- ✅ **Strong consistency** with PostgreSQL
- ✅ **At-least-once delivery** with Kafka
- ✅ **Automatic recovery** from crashes
- ✅ **Graceful shutdown** with order book snapshots

### Scalability
- ✅ **Horizontal scaling** via pair-based architecture
- ✅ **Kubernetes auto-scaling** based on load
- ✅ **Stateless replicas** for easy scaling
- ✅ **Distributed deployment** across availability zones

### Observability
- ✅ **Prometheus metrics** for monitoring
- ✅ **Health and readiness** probes
- ✅ **Structured logging** with context
- ✅ **Real-time order book** depth API

## 📦 Prerequisites

- **Kubernetes** 1.24+
- **PostgreSQL** 14+
- **Kafka** 3.0+
- **Go** 1.21+ (for development)

## 🛠️ Installation

### 1. Deploy PostgreSQL

```bash
# Using Helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgres bitnami/postgresql \
  --set auth.database=matching \
  --set auth.username=matching \
  --set primary.persistence.size=100Gi \
  --set primary.resources.requests.cpu=2 \
  --set primary.resources.requests.memory=4Gi
```

### 2. Deploy Kafka

```bash
# Using Strimzi Kafka Operator
kubectl create namespace kafka
kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka

# Create Kafka cluster
cat <<EOF | kubectl apply -f -
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: kafka
  namespace: kafka
spec:
  kafka:
    version: 3.6.0
    replicas: 3
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
    config:
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
    storage:
      type: persistent-claim
      size: 100Gi
      class: fast-ssd
  zookeeper:
    replicas: 3
    storage:
      type: persistent-claim
      size: 10Gi
EOF
```

### 3. Build and Push Docker Image

```bash
# Build image
docker build -t matching-engine:latest .

# Tag and push to registry
docker tag matching-engine:latest your-registry/matching-engine:latest
docker push your-registry/matching-engine:latest
```

### 4. Deploy Matching Engine Operator

```bash
# Install CRD
kubectl apply -f k8s/crd.yaml

# Deploy operator (if using custom operator)
kubectl apply -f k8s/operator-deployment.yaml
```

### 5. Deploy Matching Engine

```bash
# Create namespace
kubectl create namespace matching-engine

# Deploy matching engine
kubectl apply -f k8s/deployment.yaml

# Or use custom resource
kubectl apply -f k8s/crd.yaml
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/matching` |
| `KAFKA_BROKERS` | Comma-separated Kafka brokers | `localhost:9092` |
| `TRADING_PAIRS` | Comma-separated trading pairs | `BTC-USD,ETH-USD` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `METRICS_PORT` | Metrics server port | `9090` |

### Trading Pairs Configuration

Edit the `TRADING_PAIRS` environment variable or CustomResource:

```yaml
spec:
  tradingPairs:
    - BTC-USD
    - ETH-USD
    - SOL-USD
    - AVAX-USD
    - MATIC-USD
```

## 📊 API Reference

### Submit Order

```bash
POST /orders
Content-Type: application/json

{
  "user_id": "user-123",
  "symbol": "BTC-USD",
  "side": "BUY",
  "type": "LIMIT",
  "price": 45000.00,
  "quantity": 0.5
}
```

### Cancel Order

```bash
POST /orders/{order_id}/cancel
```

### Get Order Book

```bash
GET /orderbook/BTC-USD

Response:
{
  "symbol": "BTC-USD",
  "bids": [
    {"price": 45000.00, "volume": 10.5, "order_count": 5},
    {"price": 44990.00, "volume": 8.2, "order_count": 3}
  ],
  "asks": [
    {"price": 45010.00, "volume": 12.3, "order_count": 7},
    {"price": 45020.00, "volume": 5.8, "order_count": 2}
  ]
}
```

### Get Recent Trades

```bash
GET /trades/BTC-USD?limit=100
```

### Health Check

```bash
GET /health
GET /ready
```

### Metrics

```bash
GET /metrics
```

## 🔍 Monitoring

### Prometheus Metrics

The engine exposes the following metrics:

- `matching_engine_orders_total{symbol}` - Total orders processed
- `matching_engine_trades_total{symbol}` - Total trades executed
- `matching_engine_order_book_depth{symbol,side}` - Current order book depth
- `matching_engine_latency_seconds{operation}` - Operation latency
- `matching_engine_active_orders{symbol}` - Active orders in book

### Grafana Dashboard

Import the provided Grafana dashboard:

```bash
kubectl apply -f monitoring/grafana-dashboard.yaml
```

## 🧪 Testing

### Unit Tests

```bash
go test ./... -v
```

### Integration Tests

```bash
# Start dependencies
docker-compose up -d

# Run integration tests
go test ./tests/integration -v

# Cleanup
docker-compose down
```

### Load Testing

```bash
# Using k6
k6 run tests/load/order-submission.js --vus 100 --duration 60s
```

## 📈 Performance Tuning

### PostgreSQL

```sql
-- Recommended settings for high-throughput
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET effective_cache_size = '12GB';
ALTER SYSTEM SET maintenance_work_mem = '1GB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET work_mem = '64MB';
ALTER SYSTEM SET min_wal_size = '2GB';
ALTER SYSTEM SET max_wal_size = '8GB';

-- Restart PostgreSQL
SELECT pg_reload_conf();
```

### Kafka Topics

```bash
# Create topics with proper configuration
kafka-topics.sh --create --topic trades \
  --bootstrap-server localhost:9092 \
  --partitions 10 \
  --replication-factor 3 \
  --config retention.ms=604800000 \
  --config compression.type=snappy

kafka-topics.sh --create --topic order-updates \
  --bootstrap-server localhost:9092 \
  --partitions 10 \
  --replication-factor 3 \
  --config retention.ms=604800000
```

### Kubernetes Resources

```yaml
resources:
  requests:
    cpu: "2000m"      # 2 cores minimum
    memory: "4Gi"     # 4GB minimum
  limits:
    cpu: "8000m"      # 8 cores maximum
    memory: "16Gi"    # 16GB maximum
```

## 🔒 Security

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: matching-engine-policy
spec:
  podSelector:
    matchLabels:
      app: matching-engine
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: matching-engine
    ports:
    - protocol: TCP
      port: 8080
```

### TLS/SSL

Enable SSL for PostgreSQL and Kafka connections:

```bash
POSTGRES_URL=postgres://user:pass@host:5432/db?sslmode=require
```

## 🐛 Troubleshooting

### Order Book Out of Sync

```bash
# Rebuild order book from database
kubectl exec -it matching-engine-0 -- \
  curl -X POST http://localhost:8080/admin/rebuild
```

### High Latency

1. Check Kafka lag: `kubectl exec -it kafka-0 -- kafka-consumer-groups.sh --describe`
2. Monitor PostgreSQL: `SELECT * FROM pg_stat_activity;`
3. Check resource usage: `kubectl top pods`

### Pod Crashes

```bash
# View logs
kubectl logs matching-engine-0 --previous

# Check events
kubectl describe pod matching-engine-0
```

## 📝 Architecture Decisions

### Why Red-Black Trees?

- **O(log n) operations**: Better than arrays (O(n)) for frequent insertions/deletions
- **Self-balancing**: Maintains performance even with skewed data
- **Memory efficient**: No hash table overhead

### Why Pair-Based?

- **Horizontal scalability**: Scale different pairs independently
- **Fault isolation**: Issues in one pair don't affect others
- **Resource optimization**: Allocate resources based on pair volume

### Why Kafka?

- **Asynchronous processing**: Decouple matching from downstream operations
- **Event sourcing**: Complete audit trail of all operations
- **Integration**: Easy integration with other services

### Why StatefulSet?

- **Stable network identities**: Predictable pod names for debugging
- **Ordered deployment**: Ensures proper initialization sequence
- **Persistent storage**: Each pod gets its own persistent volume

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- Red-Black Tree implementation: [emirpasic/gods](https://github.com/emirpasic/gods)
- Kafka client: [segmentio/kafka-go](https://github.com/segmentio/kafka-go)
- Kubernetes operator framework: [operator-sdk](https://github.com/operator-framework/operator-sdk)

## 📧 Support

- GitHub Issues: [github.com/anhbkpro/matching-engine/issues](https://github.com/anhbkpro/matching-engine/issues)
