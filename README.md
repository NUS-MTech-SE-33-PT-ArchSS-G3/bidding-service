# Bidding Service

## About the Project

## Design

### High Level Architecture
![img.png](assets/readme/high-level-architecture.png)

### Use Cases

#### Place Bid

![img.png](assets/readme/place-bid-use-case.png)

#### Get Live Price Updates
![img.png](assets/readme/get-live-price-updates-use-case.png)

#### Get Bid History

![img.png](assets/readme/get-bid-history-simple-use-case.png)

### Interactions

#### AuctionOpened & AuctionClosed events
Subscribes to AuctionOpened and AuctionClosed events. 

On AuctionOpened the BiddingService stores the auction metadata in its cache. On AuctionClosed the BiddingService deletes the auction metadata from its cache.

![auctionopened-auctionclosed.png](assets/readme/auctionopened-auctionclosed.png)

#### Auction Service

Subscribes to PriceUpdated events. Stores snapshot for use on Auction EndAt to trigger payments

![auctionservice-priceupdated.png](assets/readme/auctionservice-priceupdated.png)

## Setting Up

### Prerequisites

### Installation
