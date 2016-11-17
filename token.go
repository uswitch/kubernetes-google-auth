package kauth

import (
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

type Token struct {
	ID     string    `json:"id"`
	Expiry time.Time `json:"expires",datastore:"expires"`
	Issued time.Time `json:"issued",datastore:"issued"`
	Email  string    `json:"user",datastore:"email"`
	UserID string    `json:"userId",datastore:"userId"`
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.Expiry)
}

func newToken(user *UserInfo) (*Token, error) {
	if !user.IsValidDomain() {
		return nil, ErrInvalidDomain
	}

	guid := uuid.NewV4()
	expires := time.Now().Add(time.Hour * 12)

	return &Token{
		ID:     guid.String(),
		Issued: time.Now(),
		Expiry: expires,
		Email:  user.Email,
		UserID: user.ID,
	}, nil
}

func tokenKey(ctx context.Context, id string) *datastore.Key {
	return datastore.NewKey(ctx, "Token", id, 0, nil)
}

func storeToken(ctx context.Context, token *Token) error {
	k := tokenKey(ctx, token.ID)
	_, err := datastore.Put(ctx, k, token)
	return err
}

func getToken(ctx context.Context, tokenID string) (*Token, error) {
	k := tokenKey(ctx, tokenID)
	var token Token
	err := datastore.Get(ctx, k, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
