package tokenstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	garbageCollectorInterval = 10 * time.Minute
)

var (
	ErrFailedToOpenTokenStore = errors.New("failed to open token store")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTtlInPast              = errors.New("TTL specified is in the past")
)

// Represents the data for a token
type Data struct {
	UserID    uuid.UUID  `json:"user_id"`
	Revoked   bool       `json:"revoked"`
	RevokedBy *string    `json:"revoked_by"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// Token store
type Store struct {
	db     *badger.DB
	cancel context.CancelFunc
}

// Create a new token store
func NewStore(ctx context.Context, dataPath string, zLog zerolog.Logger) (*Store, error) {
	opts := badger.DefaultOptions(dataPath).WithLogger(&logger{log: zLog})
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToOpenTokenStore, err)
	}

	ctx, cancel := context.WithCancel(ctx)
	store := &Store{
		db:     db,
		cancel: cancel,
	}

	go store.runGC(ctx)

	return store, nil
}

// Retrieves the token data for a given JTI
func (s *Store) GetToken(ctx context.Context, jti string) (*Data, error) {
	var data Data
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(jti))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &data)
		})
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	return &data, nil
}

// Store a token with a TTL equal to its expiration time
func (s *Store) StoreToken(ctx context.Context, jti string, userID uuid.UUID, expiration time.Duration) error {
	data := Data{
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		Revoked:   false,
		ExpiresAt: time.Now().UTC().Add(expiration),
	}

	val, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte(jti), val).WithTTL(expiration))
	})
}

// Updates a token value
func (s *Store) UpdateToken(ctx context.Context, jti string, data *Data) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}

	remainingTTL := time.Until(data.ExpiresAt)

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte(jti), val).WithTTL(remainingTTL))
	})
}

// Stops the garbage collection and closes the database
func (s *Store) Close() error {
	s.cancel()
	return s.db.Close()
}

// Runs the BadgerDB value log garbage collection
func (s *Store) runGC(ctx context.Context) {
	ticker := time.NewTicker(garbageCollectorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				if err := s.db.RunValueLogGC(0.5); err != nil {
					break
				}
			}
		}
	}
}
