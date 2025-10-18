# DefaultApi

All URIs are relative to *http://localhost:8083*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**apiV1BidsAuctionIdGet**](#apiv1bidsauctionidget) | **GET** /api/v1/bids/{auctionId} | List bids for an auction|

# **apiV1BidsAuctionIdGet**
> ListBidsResponse apiV1BidsAuctionIdGet()

Returns bids for a given auction, newest first by default, using cursor pagination. `cursor` is from the previous page\'s (`nextCursor`). 

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from '@kei/bid-query-api-client';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let auctionId: string; //The auction ID (default to undefined)
let cursor: string; //Cursor from the previous page, omit for the first page (optional) (default to undefined)
let limit: number; //Page size. Max 200. (optional) (default to 50)
let direction: 'desc' | 'asc'; //Sort in descending or ascending order. (optional) (default to 'desc')

const { status, data } = await apiInstance.apiV1BidsAuctionIdGet(
    auctionId,
    cursor,
    limit,
    direction
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **auctionId** | [**string**] | The auction ID | defaults to undefined|
| **cursor** | [**string**] | Cursor from the previous page, omit for the first page | (optional) defaults to undefined|
| **limit** | [**number**] | Page size. Max 200. | (optional) defaults to 50|
| **direction** | [**&#39;desc&#39; | &#39;asc&#39;**]**Array<&#39;desc&#39; &#124; &#39;asc&#39;>** | Sort in descending or ascending order. | (optional) defaults to 'desc'|


### Return type

**ListBidsResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | A page of bids. |  * X-Next-Cursor - Same value as &#x60;nextCursor&#x60; in the body for convenience <br>  * X-Request-Id - Echoes back the request ID, if not provided by the client, server generates one. <br>  |
|**400** | Bad request (invalid cursor, params, etc.) |  -  |
|**404** | Auction not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

