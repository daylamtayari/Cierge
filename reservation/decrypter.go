package reservation

import "context"

type Decrypter interface {
	Decrypt(ctx context.Context, encrypted []byte) (string, error)
}
