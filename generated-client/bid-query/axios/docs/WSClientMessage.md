# WSClientMessage


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**type** | **string** |  | [default to undefined]
**ids** | **Array&lt;string&gt;** |  | [default to undefined]
**resumeFrom** | **{ [key: string]: string; }** | Map of auctionId -&gt; resumeToken (last seen version/seq). | [optional] [default to undefined]
**ts** | **string** |  | [default to undefined]

## Example

```typescript
import { WSClientMessage } from '@kei/bid-query-api-client';

const instance: WSClientMessage = {
    type,
    ids,
    resumeFrom,
    ts,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
