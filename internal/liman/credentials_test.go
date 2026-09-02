package liman

import (
	"errors"
	"testing"

	"github.com/limanmys/render-engine/app/models"
	"gorm.io/gorm"
)

type fakeServerKeyRepository struct {
	userKey       *models.ServerKey
	userErr       error
	sharedKey     *models.ServerKey
	sharedErr     error
	sharedLookups int
}

func (r *fakeServerKeyRepository) FindUserKey(_, _ string) (*models.ServerKey, error) {
	return r.userKey, r.userErr
}

func (r *fakeServerKeyRepository) FindSharedKey(_ string) (*models.ServerKey, error) {
	r.sharedLookups++
	return r.sharedKey, r.sharedErr
}

func TestResolveServerKeyPrefersPersonalKey(t *testing.T) {
	repository := &fakeServerKeyRepository{
		userKey:   &models.ServerKey{ID: "personal", UserID: "requesting-user"},
		sharedKey: &models.ServerKey{ID: "shared", UserID: "key-owner", Shared: true},
	}

	key, err := resolveServerKey(repository, "requesting-user", "server")
	if err != nil {
		t.Fatalf("resolveServerKey returned an error: %v", err)
	}
	if key.ID != "personal" {
		t.Fatalf("expected personal key, got %q", key.ID)
	}
	if repository.sharedLookups != 0 {
		t.Fatalf("shared key was queried even though a personal key exists")
	}
}

func TestResolveServerKeyFallsBackToExplicitlySharedKey(t *testing.T) {
	repository := &fakeServerKeyRepository{
		userErr:   gorm.ErrRecordNotFound,
		sharedKey: &models.ServerKey{ID: "shared", UserID: "key-owner", Shared: true},
	}

	key, err := resolveServerKey(repository, "requesting-user", "server")
	if err != nil {
		t.Fatalf("resolveServerKey returned an error: %v", err)
	}
	if key.ID != "shared" || !key.Shared {
		t.Fatalf("expected explicitly shared key, got %#v", key)
	}
}

func TestResolveServerKeyFailsClosedWhenNoSharedKeyExists(t *testing.T) {
	repository := &fakeServerKeyRepository{
		userErr:   gorm.ErrRecordNotFound,
		sharedErr: gorm.ErrRecordNotFound,
	}

	_, err := resolveServerKey(repository, "requesting-user", "server")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record-not-found, got %v", err)
	}
}

func TestResolveServerKeyDoesNotHideDatabaseErrors(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	repository := &fakeServerKeyRepository{
		userErr:   databaseErr,
		sharedKey: &models.ServerKey{ID: "shared", Shared: true},
	}

	_, err := resolveServerKey(repository, "requesting-user", "server")
	if !errors.Is(err, databaseErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if repository.sharedLookups != 0 {
		t.Fatalf("shared fallback must not run after a database error")
	}
}
