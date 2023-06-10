# sovos cli
A small CLI tool to allow interaction with the API for the Portugese tax system Sovos.
***

## env

The .env.example file contains all the necessary endpoint names already, but will require your own API key, secret and User ID (which is generally your Portugese VAT number preceeded by PT.  e.g. `PT123456789`)

## commands
`-send {record_type} '{json_string}'`

the send command will prepare and send the record to the Sovos receiving queue ready to be processed.  At present only the `invoice` record type is available.

it also takes in a minified json string enclosed in quotes.  the struct in the repo shows the fields mapped and can be customised, but this is the minimum required fields for a record to be accepted in the Sovos system (example below)

&nbsp;

`-getmessages`

the `getmessages` flag will retrieve all messages on the currently selecte Sovos queue (defined by your environment variable).  Each message will be in the format of MessageId, Sender and Receiver
```JSON
{
    "MessageId": "20230101121212.e2e45465-c6aa-4737-864c-f1ff0309753b@l-qa-fes60",
    "Receiver": "123456789",
    "Sender": "565656565656"
},
```

&nbsp;

`-getmessage {message_id}`

passing a message ID (like the one above) to this flag will attempt to retrieve the details of that specific message.  You are given the option of either returning the raw payload:
```JSON
{
  "MessageId": "20230101121212.e2e45465-c6aa-4737-864c-f1ff0309753b@l-qa-fes60",
  "Receiver": "123456789",
  "Sender": "565656565656",
  "Base64Data": "base64_encoded_result_goes_here"
}
```
or just the decoded data inside the `Base64Data` field which will be an XML structured result detailing any errors with the data that was sent or any successful response IDs

&nbsp;

`-processmessage {message_id}`

passing the message ID to this flag will indicate to Sovos that the message has been retrieved and is safe to remove from their queue.  Only use this once you have retrieved the data inside the message as once it is removed from the queue you're unable to get it back

&nbsp;

`-clearmessages`

this flag will attempt to clear all outstanding messages on the current queue.  You are given a warning and notice of which queue will be cleared.  Choosing to continue will display each message that is cleared as they are set to processed.

&nbsp;

`-decode 'base_64_string_goes_here'`

this is a deprecated flag (now included as an option in the `getmessage` flag) but is left here in case it's useful.  pass a base64 encoded string (in quotes) here and get the decoded version output to the terminal

***
## example invoice payload
```JSON
{
    "invoice":
    {
        "@_documentCorrelationId": "00000000-0000-0000-0000-000123456789",
        "@_docTypeId": "47",
        "@_docInstanceId": 123456789,
        "@_docPlatform": "0",
        "@_serie": "INV",
        "currencyISOCode": "EUR",
        "documentReferences":
        {
            "thirdPartyErpInternalReference": "INV0123456789"
        },
        "documentDates":
        {
            "documentDate": "2023-01-01",
            "goodsServiceAvailableDate": "2023-01-01",
            "dueDate": "2023-01-01"
        },
        "partyInformation":
        {
            "seller":
            {
                "name": "Example Seller Ltd",
                "country": "PT",
                "vatNumber": "123456789",
                "address": "Example Seller Address",
                "city": "ExampleCity",
                "zipArea": "ExampleZipArea",
                "zipCode": "1234-567",
                "companyRegistrationNumber": "0123456",
                "companyRegistrationLocation": "United Kingdom",
                "socialCapitalValue": "1000000"
            },
            "buyer":
            {
                "name": "Example Customer",
                "email": "example.customer@example-email.com",
                "address": "Example Customer Address",
                "zipCode": "9876-5433",
                "zipArea": "ExampleZipArea",
                "country": "PT",
                "vatNumber": "999999990",
                "isFinalConsumer": true
            }
        },
        "lineItem":
        [
            {
                "@_number": 1,
                "sellerAssignedTradeItemIdentification": "SKU-ABC",
                "itemDescription": "Example Product 1",
                "netPrice": 26.83,
                "netLineAmount": 26.83,
                "grossPrice": 26.83,
                "grossLineAmount": 26.83,
                "lineTotalPayableAmount": 33,
                "quantity":
                {
                    "value": 1,
                    "unitCodeValue": "UNIT"
                },
                "lineVat":
                {
                    "taxableAmount": 26.83,
                    "taxPercentage": 23,
                    "taxTotalValue": 6.17
                }
            },
            {
                "@_number": 2,
                "sellerAssignedTradeItemIdentification": "SKU-DEF",
                "itemDescription": "Example Product 2",
                "netPrice": 16.26,
                "netLineAmount": 16.26,
                "grossPrice": 16.26,
                "grossLineAmount": 16.26,
                "lineTotalPayableAmount": 20,
                "quantity":
                {
                    "value": 1,
                    "unitCodeValue": "UNIT"
                },
                "lineVat":
                {
                    "taxableAmount": 16.26,
                    "taxPercentage": 23,
                    "taxTotalValue": 3.74
                }
            },
            {
                "@_number": 3,
                "sellerAssignedTradeItemIdentification": "SKU-XYZ",
                "itemDescription": "Example Product 3",
                "netPrice": 31.71,
                "netLineAmount": 31.71,
                "grossPrice": 31.71,
                "grossLineAmount": 31.71,
                "lineTotalPayableAmount": 39,
                "quantity":
                {
                    "value": 1,
                    "unitCodeValue": "UNIT"
                },
                "lineVat":
                {
                    "taxableAmount": 31.71,
                    "taxPercentage": 23,
                    "taxTotalValue": 7.29
                }
            },
            {
                "@_number": 4,
                "sellerAssignedTradeItemIdentification": "Shipping Cost",
                "itemDescription": "Shipping Cost",
                "netPrice": 0,
                "netLineAmount": 0,
                "grossPrice": 0,
                "grossLineAmount": 0,
                "lineTotalPayableAmount": 0,
                "quantity":
                {
                    "value": 1,
                    "unitCodeValue": "UNIT"
                },
                "lineVat":
                {
                    "taxableAmount": 0,
                    "taxPercentage": 23,
                    "taxTotalValue": 0
                }
            }
        ],
        "documentTotals":
        {
            "numberOfLines": 4,
            "totalAmountPayable": 82.8,
            "totalVatTaxableAmount": 67.32,
            "totalVatAmount": 15.48,
            "totalGrossAmount": 67.32,
            "totalNetAmount": 67.32,
            "vatSummary":
            {
                "taxPercentage": 23,
                "taxTotalValue": 15.48,
                "taxableAmount": 67.32
            }
        },
        "emailNotification":
        {
            "email": "example.customer@example-email.com",
            "languageCode": "PT"
        }
    }
} 
