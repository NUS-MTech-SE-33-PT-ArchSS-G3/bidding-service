## @kei/bid-query-api-client@1.0.0

This generator creates TypeScript/JavaScript client that utilizes [axios](https://github.com/axios/axios). The generated Node module can be used in the following environments:

Environment
* Node.js
* Webpack
* Browserify

Language level
* ES5 - you must have a Promises/A+ library installed
* ES6

Module system
* CommonJS
* ES6 module system

It can be used in both TypeScript and JavaScript. In TypeScript, the definition will be automatically resolved via `package.json`. ([Reference](https://www.typescriptlang.org/docs/handbook/declaration-files/consumption.html))

### Building

To build and compile the typescript sources to javascript use:
```
npm install
npm run build
```

### Publishing

First build the package then run `npm publish`

### Consuming

navigate to the folder of your consuming project and run one of the following commands.

_published:_

```
npm install @kei/bid-query-api-client@1.0.0 --save
```

_unPublished (not recommended):_

```
npm install PATH_TO_GENERATED_PACKAGE --save
```

### Documentation for API Endpoints

All URIs are relative to *http://localhost:8083*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*DefaultApi* | [**auctionsAuctionIdBidsGet**](docs/DefaultApi.md#auctionsauctionidbidsget) | **GET** /auctions/{auctionId}/bids | List bids for an auction


### Documentation For Models

 - [Ack](docs/Ack.md)
 - [Bid](docs/Bid.md)
 - [ErrorFrame](docs/ErrorFrame.md)
 - [ListBidsResponse](docs/ListBidsResponse.md)
 - [Ping](docs/Ping.md)
 - [Pong](docs/Pong.md)
 - [PriceEvent](docs/PriceEvent.md)
 - [ProblemDetails](docs/ProblemDetails.md)
 - [Subscribe](docs/Subscribe.md)
 - [Unsubscribe](docs/Unsubscribe.md)
 - [WSClientMessage](docs/WSClientMessage.md)
 - [WSServerMessage](docs/WSServerMessage.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization


Authentication schemes defined for the API:
<a id="bearerAuth"></a>
### bearerAuth

- **Type**: Bearer authentication (JWT)

