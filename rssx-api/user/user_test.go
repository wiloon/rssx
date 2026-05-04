package user

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"rssx/common"
)

func TestMain(m *testing.M) {
	common.InitForTesting()
	os.Exit(m.Run())
}

// TestUser_Register_GeneratesUUID verifies that Register() sets a non-empty UUID
// using uuid.New() (changed from uuid.NewV4() in this PR).
func TestUser_Register_GeneratesUUID(t *testing.T) {
	u := &User{Name: "reg_uuid_test", Password: "hashvalue"}
	u.Register()

	if u.Id == "" {
		t.Error("Register() must set a non-empty UUID for Id")
	}
	if u.CreateTime == "" {
		t.Error("Register() must set CreateTime")
	}
}

// TestUser_Register_UniqueIDs verifies each Register call produces a distinct UUID.
func TestUser_Register_UniqueIDs(t *testing.T) {
	u1 := &User{Name: "unique_id_user1", Password: "pw"}
	u2 := &User{Name: "unique_id_user2", Password: "pw"}
	u1.Register()
	u2.Register()

	if u1.Id == u2.Id {
		t.Errorf("expected distinct UUIDs, both got %s", u1.Id)
	}
}

// TestUser_IsExist_True verifies IsExist() returns true after Register().
func TestUser_IsExist_True(t *testing.T) {
	u := &User{Name: "isexist_true_user", Password: "pw"}
	u.Register()

	check := &User{Name: "isexist_true_user"}
	if !check.IsExist() {
		t.Error("IsExist() should return true for a registered user")
	}
}

// TestUser_IsExist_False verifies IsExist() returns false for an unknown user.
func TestUser_IsExist_False(t *testing.T) {
	check := &User{Name: "definitely_not_registered_xyzabc"}
	if check.IsExist() {
		t.Error("IsExist() should return false for a user that has never been registered")
	}
}

// TestUser_Validate_CopiesId is a regression test for the PR fix:
// Before the fix, Validate() checked the password but never copied tmp.Id back to u.Id,
// leaving every issued JWT with an empty "id" claim.
// After the fix: u.Id = tmp.Id is set on successful validation.
func TestUser_Validate_CopiesId(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}
	registered := &User{Name: "validate_id_copy_user", Password: string(hash)}
	registered.Register()
	registeredID := registered.Id

	v := &User{Name: "validate_id_copy_user", Password: "secret123"}
	if !v.Validate() {
		t.Fatal("Validate() returned false for correct credentials")
	}
	if v.Id != registeredID {
		t.Errorf("Validate() must copy DB Id to u.Id: expected %s, got %q", registeredID, v.Id)
	}
}

// TestUser_Validate_WrongPassword verifies Validate() returns false for a bad password.
func TestUser_Validate_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	u := &User{Name: "validate_wrong_pw_user", Password: string(hash)}
	u.Register()

	v := &User{Name: "validate_wrong_pw_user", Password: "wrongpass"}
	if v.Validate() {
		t.Error("Validate() should return false for wrong password")
	}
}

// TestUser_Validate_UnknownUser verifies Validate() returns false for a non-existent user.
func TestUser_Validate_UnknownUser(t *testing.T) {
	v := &User{Name: "ghost_user_nonexistent", Password: "anypassword"}
	if v.Validate() {
		t.Error("Validate() should return false for a user that does not exist")
	}
}
