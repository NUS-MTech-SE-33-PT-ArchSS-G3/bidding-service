# DefaultApi

All URIs are relative to *http://localhost:8081*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**auctionsAuctionIdBidsPost**](#auctionsauctionidbidspost) | **POST** /auctions/{auctionId}/bids | Place a bid on an auction|

# **auctionsAuctionIdBidsPost**
> PlaceBidResponse auctionsAuctionIdBidsPost(placeBidRequest)

Place a new bid for a given auction.   

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    PlaceBidRequest
} from '@kei/bid-command-api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let auctionId: string; //ID of the auction to bid on (default to undefined)
let placeBidRequest: PlaceBidRequest; //

const { status, data } = await apiInstance.auctionsAuctionIdBidsPost(
    auctionId,
    placeBidRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **placeBidRequest** | **PlaceBidRequest**|  | |
| **auctionId** | [**string**] | ID of the auction to bid on | defaults to undefined|


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

