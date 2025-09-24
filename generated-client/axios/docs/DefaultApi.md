# DefaultApi

All URIs are relative to *http://localhost:8080*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**auctionsAuctionIdBidsPost**](#auctionsauctionidbidspost) | **POST** /auctions/{auctionId}/bids | Place a bid on an auction|

# **auctionsAuctionIdBidsPost**
> PlaceBidResponse auctionsAuctionIdBidsPost(placeBidRequest)

Place a new bid for a given auction.   Requires a valid JWT token.   Uses an Idempotency-Key header to ensure retries are safe. 

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    PlaceBidRequest
} from '@kei/bidding-api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let auctionId: string; //ID of the auction to bid on. (default to undefined)
let idempotencyKey: string; //Unique key to prevent duplicate bid submissions (default to undefined)
let placeBidRequest: PlaceBidRequest; //

const { status, data } = await apiInstance.auctionsAuctionIdBidsPost(
    auctionId,
    idempotencyKey,
    placeBidRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **placeBidRequest** | **PlaceBidRequest**|  | |
| **auctionId** | [**string**] | ID of the auction to bid on. | defaults to undefined|
| **idempotencyKey** | [**string**] | Unique key to prevent duplicate bid submissions | defaults to undefined|


### Return type

**PlaceBidResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Bid accepted |  -  |
|**400** | Invalid request (e.g., missing fields) |  -  |
|**401** | Unauthorized (missing or invalid JWT) |  -  |
|**409** | Conflict (out-of-date version or duplicate) |  -  |
|**422** | Bid rejected (below minimum increment, auction closed, etc.) |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

