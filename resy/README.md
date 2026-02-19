# Resy API Library

While not feature complete, this API library covers 90+% of things you would ever need when interacting with the Resy API. Similarly, it covers the overwhelming majority of fields are mapped in structs, not every single field is simply due to how verbose the API is and how much irrelevant data exists.

If there are any fields or methods you want to see implemented, reach out and I will happily implement them.

## Usage

Create a new client by calling the `NewClient` function. This accepts the following parameters:
- An `http.Client` that will be used as the underlying HTTP client that is used to make requests, otherwise a new `http.Client` is used.
- `Tokens` containing the API key and optionally the authenticated user tokens. If no API key is specified, the default will be used.
- A string for the user agent to be used. If an empty string is specified, a default value representing a generic popular user agent will be used. Resy's API requires a user agent to be present.

With the `Client`, you can then use it to call any of the methods provided by this library.

## Tests

Tests were created to verify and track the functioning of this API library.

If multiple tests are ran at once, they will often fail due to rate limiting.

### Environment Variables

The following environment variables can be used to set values for tests:
- `RESY_AUTH_TOKEN` - User auth token (required for tests of endpoints that require authentication)
- `RESY_API_KEY` - Override the default API key
- `RESY_TEST_DATE` - Future date for slot tests (default: 14 days from now)
- `RESY_TEST_VENUE_ID` - Venue ID for tests (default: 54602)
- `REST_TEST_RESERVATION_TOKEN` - Reservation token for reservation modification tests
- `RESY_ENABLE_BOOKING_TESTS` - Enable dangerous booking tests

## Understanding the Resy API

Certain core concepts of the Resy API are outlined below:

### Authentication

Resy's API has both 'unauthenticated' requests for generic actions such as searching restaurants, and authenticated actions such as for booking a reservation.

#### 'Unauthenticated'

Requests for generic actions that are for all intents and purposes unauthenticated, still require an API key value. This API key value is sourced from the application JavaScript of the Resy web application however, it has for at least the last 5 years (and likely longer) been a static value. 

This static value is stored in the `DefaultApiKey` variable but if you prefer to fetch it dynamically or it ever changes, the logic to retrieve the current API key server by the web application is present in the `FetchApiKey` function. Worth noting that the `FetchApiKey` function is heavy as it requires retrieving the index and then a multi-MB JavaScript file and parsing it to retrieve the API key.

At the time of writing, this API key is never scoped to a particular user.

#### Authenticated

When authenticating as a user, Resy returns an auth token valid for 45 days and a refresh token valid for 90 days.

Only the auth token needs to be provided in authenticated requests, specified in both the `X-Resy-Auth-Token` and `X-Resy-Universal-Auth` request headers.

The refresh token can be used to obtain a new authentication and refresh token each with an expiry of another 45 and 90 days respectively.

### Reservation

The process for acquiring a reservation contains three core stages:
- Slot retrieval
- Slot details
- Booking

When specifying the venue, date, and party size, Resy will return all available slots and a token corresponding to that slot. The token is just a string that contains the venue ID, service ID, date, time slot, party size, etc., it is not a secret or unique token.

Using this slot token, you can then get the details of a reservation slot. This will provide you with a booking token.

This booking token can then be used to perform the actual booking of the reservation. If the reservation requires a deposit or fee, a payment method ID correlating to a payment method of the user, must be specified. On successful completion, this will return your reservation token that can be used to manage the reservation and the ID of the reservation.

This flow and the required fields is illustrated below:

```
  ┌─────────────────────────────────────────────────────┐
  │                    1. GET USER                      │
  │                    GET /2/user                      │
  ├─────────────────────────────────────────────────────┤
  │ Request: (none)                                     │
  │ Response: Payment Method ID                         │
  └─────────────────────────────────────────────────────┘
                           ↓
  ┌─────────────────────────────────────────────────────┐
  │                   2. GET SLOTS                      │
  │                   POST /4/find                      │
  ├─────────────────────────────────────────────────────┤
  │ Request: Venue ID, Reservation Date, Party Size     │
  │ Response: Slot Token, Slot Time                     │
  └─────────────────────────────────────────────────────┘
                           ↓
  ┌─────────────────────────────────────────────────────┐
  │                3. GET SLOT DETAILS                  │
  │                 POST /3/details                     │
  ├─────────────────────────────────────────────────────┤
  │ Request: Slot Token, Reservation Date, Party Size   │
  │ Response: Booking Token                             │
  └─────────────────────────────────────────────────────┘
                           ↓
  ┌─────────────────────────────────────────────────────┐
  │                4. BOOK RESERVATION                  │
  │                   POST /3/book                      │
  ├─────────────────────────────────────────────────────┤
  │ Request: Booking Token, Payment Method ID (opt)     │
  │ Response: Reservation Token, Reservation ID         │
  └─────────────────────────────────────────────────────┘

```
