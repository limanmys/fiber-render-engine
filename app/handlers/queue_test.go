package handlers

import "testing"

func TestBindQueueUserUsesAuthenticatedIdentity(t *testing.T) {
	data := map[string]interface{}{}

	if err := bindQueueUser(data, "authenticated-user"); err != nil {
		t.Fatalf("bindQueueUser returned an error: %v", err)
	}
	if data["user_id"] != "authenticated-user" {
		t.Fatalf("expected authenticated user id, got %#v", data["user_id"])
	}
}

func TestBindQueueUserRejectsDifferentSuppliedIdentity(t *testing.T) {
	data := map[string]interface{}{"user_id": "other-user"}

	if err := bindQueueUser(data, "authenticated-user"); err == nil {
		t.Fatal("expected mismatched queue user to be rejected")
	}
	if data["user_id"] != "other-user" {
		t.Fatalf("mismatched identity must not be rewritten before rejection")
	}
}

func TestBindQueueUserPreservesMatchingIdentity(t *testing.T) {
	data := map[string]interface{}{"user_id": "authenticated-user"}

	if err := bindQueueUser(data, "authenticated-user"); err != nil {
		t.Fatalf("matching queue user was rejected: %v", err)
	}
}
