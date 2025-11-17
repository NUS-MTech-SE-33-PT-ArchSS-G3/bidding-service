# Bidding Services

## About the Project

## Project Structure
```
├── docker-compose.yaml
├── generated-client
├── grafana
├── openapi
│   ├── bid-command
│   └── bid-query
├── pkg
├── prometheus.yml
├── qodana.yaml
├── scripts
├── services
│   ├── auction-projector
│   │   ├── Dockerfile
│   │   ├── cmd
│   │   └── internal
│   │       ├── events
│   │       ├── projections
│   │       └── projector
│   ├── bid-command
│   │   ├── Dockerfile
│   │   ├── cmd
│   │   ├── internal
│   │   │   ├── application
│   │   │   ├── domain
│   │   │   ├── infrastructure
│   │   │   ├── presentation
│   │   │   └── server
│   │   ├── openapi
│   │   └── sqlc
│   ├── bid-projector
│   │   ├── Dockerfile
│   │   ├── cmd
│   │   └── internal
│   │       ├── events
│   │       ├── projections
│   │       └── projector
│   ├── bid-query
│   │   ├── Dockerfile
│   │   ├── cmd
│   │   ├── internal
│   │   │   ├── application
│   │   │   ├── infrastructure
│   │   │   ├── presentation
│   │   │   ├── read_model
│   │   │   └── server
│   │   └── openapi
└── sqlc.yaml
```
## Design

### High Level Architecture
![img.png](assets/readme/high-level-architecture.png)

### Use Cases

[//]: # ()
[//]: # (#### Place Bid)

[//]: # ()
[//]: # (![img.png]&#40;assets/readme/place-bid-use-case.png&#41;)

[//]: # ()
[//]: # (#### Get Live Price Updates)

[//]: # (![img.png]&#40;assets/readme/get-live-price-updates-use-case.png&#41;)

[//]: # ()
[//]: # (#### Get Bid History)

[//]: # ()
[//]: # (![img.png]&#40;assets/readme/get-bid-history-simple-use-case.png&#41;)

[//]: # ()
[//]: # (### Interactions)

[//]: # ()
[//]: # (#### AuctionOpened & AuctionClosed events)

[//]: # (Subscribes to AuctionOpened and AuctionClosed events. )

[//]: # ()
[//]: # (On AuctionOpened the BiddingService stores the auction metadata in its cache. On AuctionClosed the BiddingService deletes the auction metadata from its cache.)

[//]: # ()
[//]: # (![auctionopened-auctionclosed.png]&#40;assets/readme/auctionopened-auctionclosed.png&#41;)

[//]: # ()
[//]: # (#### Auction Service)

[//]: # ()
[//]: # (Subscribes to PriceUpdated events. Stores snapshot for use on Auction EndAt to trigger payments)

[//]: # ()
[//]: # (![auctionservice-priceupdated.png]&#40;assets/readme/auctionservice-priceupdated.png&#41;)

## Links
- [Kafka UI](http://localhost:8080)
- [Bid Command Swagger](http://localhost:8082/swagger/openapi/)
- [Bid Query Swagger](http://localhost:8083/swagger/openapi/)
- [Prometheus](http://localhost:9090)
- [Grafana Bid Platform Dashboard](http://localhost:3000/d/bid-platform/bid-platform-overview)
    - Username: admin
    - Password: admin

## Run Locally

```bash
open -a Docker
docker compose up -d 
```

## pprof

go tool pprof -http=:8081 http://127.0.0.1:6062/debug/pprof/profile
go tool pprof -http=:8081 http://127.0.0.1:6062/debug/pprof/heap
curl http://127.0.0.1:6062/debug/pprof/goroutine?debug=1

## dev

```bash
docker exec -it kafka bash

# open
kafka-console-producer.sh --bootstrap-server kafka:9092 \
--topic auction.opened --property parse.key=true --property key.separator=":"

a_seeded:{"auctionId":"a_seeded","endsAt":"2025-10-16T02:45:00Z","startingPrice":100,"minIncrement":10,"version":1}
b_seeded:{"auctionId":"b_seeded","endsAt":"2025-10-16T02:45:00Z","startingPrice":100,"minIncrement":10,"version":1}

# close
kafka-console-producer.sh --bootstrap-server kafka:9092 \
--topic auction.closed --property parse.key=true --property key.separator=":"
a_seeded:{"auctionId":"a_seeded","closedAt":"2025-10-16T03:00:00Z","version":2}
```

```bash
curl -X 'POST' \
  'http://localhost:8082/api/v1/bids/a_seeded' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "amount": 131.5,
  "bidderId": "user_123"
}'
```

```bash
docker exec -it kafka bash

# open
kafka-console-producer.sh --bootstrap-server kafka:9092 \
--topic auction.opened --property parse.key=true --property key.separator=":"

b_seeded:{"auctionId":"b_seeded","endsAt":"2025-10-16T02:45:00Z","startingPrice":100,"minIncrement":10,"version":1}
```

docker compose up -d --build bid-command bid-query 