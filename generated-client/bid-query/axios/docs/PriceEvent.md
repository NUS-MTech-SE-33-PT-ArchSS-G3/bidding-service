# PriceEvent


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**type** | **string** |  | [default to undefined]
**auctionId** | **string** |  | [default to undefined]
**currentPrice** | **number** |  | [default to undefined]
**minNextBid** | **number** |  | [optional] [default to undefined]
**version** | **number** | Monotonic version/seq for resume. | [default to undefined]
**resumeToken** | **string** | Opaque token (e.g., same as version or compound). | [optional] [default to undefined]
**at** | **string** |  | [default to undefined]

## Example

```typescript
import { PriceEvent } from '@kei/bid-query-api-client';

const instance: PriceEvent = {
    type,
    auctionId,
    currentPrice,
    minNextBid,
    version,
    resumeToken,
    at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
