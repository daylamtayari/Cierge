# Cierge API Library

Complete API library for the Cierge API.

## Usage

Create a new client by calling the `NewClient` function. This accepts the following parameters:
- An `http.Client` that will be used as the underlying HTTP client that is used to make requests, otherwise a new `http.Client` is used.
- URL of the Cierge server
- User's API key that is used for authentication

With the `Client`, you can then use it to call any of the methods provided by this library.
