package model_test

import (
	"testing"

	"gorsk-gin/internal"
)

func TestUpdateLastLogin(t *testing.T) {
	user := &model.User{
		FirstName: "TestGuy",
	}
	user.UpdateLastLogin()
	if user.LastLogin.IsZero() {
		t.Errorf("Last login time was not changed")
	}
}
