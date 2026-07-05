package model

import "testing"

func TestRoleValid(t *testing.T) {
	valid := []Role{RoleBuyer, RoleOrganizer, RoleAdmin}
	for _, r := range valid {
		if !r.Valid() {
			t.Errorf("Role(%q).Valid() = false; want true", r)
		}
	}
	if Role("wizard").Valid() {
		t.Error("Role(\"wizard\").Valid() = true; want false")
	}
	if Role("").Valid() {
		t.Error("empty Role.Valid() = true; want false")
	}
}

func TestDefaultRoleIsBuyer(t *testing.T) {
	if RoleBuyer != "buyer" {
		t.Errorf("RoleBuyer = %q; want buyer", RoleBuyer)
	}
}

func TestTableNames(t *testing.T) {
	if (User{}).TableName() != "users" {
		t.Errorf("User table = %q; want users", (User{}).TableName())
	}
	if (RefreshToken{}).TableName() != "refresh_tokens" {
		t.Errorf("RefreshToken table = %q; want refresh_tokens", (RefreshToken{}).TableName())
	}
}
