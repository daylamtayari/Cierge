// Go library for the Resy API
//
// The Resy API is *incredibly* verbose and as such, not all fields are
// handled by the package and this should not be expected to be to be
// feature complete of the Resy API.
// If there are any structs or fields that are not handled that you desire,
// please submit a PR and I will happily implement them.
package resy

// Resy API host
const Host = "https://api.resy.com"

type Tokens struct {
	ApiKey string
	Token  string
}
