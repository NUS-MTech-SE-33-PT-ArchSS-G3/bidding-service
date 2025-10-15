# WSServerMessage


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
**op** | **string** |  | [default to undefined]
**ids** | **Array&lt;string&gt;** |  | [default to undefined]
**code** | **string** |  | [default to undefined]
**message** | **string** |  | [default to undefined]
**details** | **{ [key: string]: any; }** |  | [optional] [default to undefined]
**ts** | **string** |  | [default to undefined]

## Example

```typescript
import { WSServerMessage } from '@kei/bid-query-api-client';

const instance: WSServerMessage = {
    type,
    auctionId,
    currentPrice,
    minNextBid,
    version,
    resumeToken,
    at,
    op,
    ids,
    code,
    message,
    details,
    ts,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
