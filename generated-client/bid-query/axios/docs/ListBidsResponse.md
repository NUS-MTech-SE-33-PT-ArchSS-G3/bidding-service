# ListBidsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**Array&lt;Bid&gt;**](Bid.md) |  | [default to undefined]
**nextCursor** | **string** | Opaque token for the next page, or null if no more results. | [optional] [default to undefined]
**hasMore** | **boolean** |  | [optional] [default to undefined]

## Example

```typescript
import { ListBidsResponse } from '@kei/bid-query-api-client';

const instance: ListBidsResponse = {
    items,
    nextCursor,
    hasMore,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
