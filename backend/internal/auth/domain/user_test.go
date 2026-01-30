package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestUser_FullName(t *testing.T) {
	u := User{
		FirstName: "John",
		LastName:  "Doe",
	}

	if got := u.FullName(); got != "John Doe" {
		t.Errorf("FullName() = %q, want %q", got, "John Doe")
	}
}

func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		want     bool
	}{
		{"active user", true, true},
		{"inactive user", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{IsActive: tt.isActive}
			if got := u.CanLogin(); got != tt.want {
				t.Errorf("CanLogin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_HasTenant(t *testing.T) {
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	tenantID3 := uuid.New()

	u := User{
		TenantRoles: []UserTenantRole{
			{TenantID: tenantID1, Role: RoleManager},
			{TenantID: tenantID2, Role: RoleWaiter},
		},
	}

	if !u.HasTenant(tenantID1) {
		t.Error("Should have tenant 1")
	}

	if !u.HasTenant(tenantID2) {
		t.Error("Should have tenant 2")
	}

	if u.HasTenant(tenantID3) {
		t.Error("Should not have tenant 3")
	}
}

func TestUser_GetRoleForTenant(t *testing.T) {
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	tenantID3 := uuid.New()

	u := User{
		TenantRoles: []UserTenantRole{
			{TenantID: tenantID1, Role: RoleManager},
			{TenantID: tenantID2, Role: RoleWaiter},
		},
	}

	if got := u.GetRoleForTenant(tenantID1); got != RoleManager {
		t.Errorf("GetRoleForTenant(tenant1) = %q, want %q", got, RoleManager)
	}

	if got := u.GetRoleForTenant(tenantID2); got != RoleWaiter {
		t.Errorf("GetRoleForTenant(tenant2) = %q, want %q", got, RoleWaiter)
	}

	if got := u.GetRoleForTenant(tenantID3); got != "" {
		t.Errorf("GetRoleForTenant(unknown) = %q, want empty", got)
	}
}

func TestUser_TenantCount(t *testing.T) {
	u1 := User{TenantRoles: []UserTenantRole{}}
	if u1.TenantCount() != 0 {
		t.Error("Empty TenantRoles should have count 0")
	}

	u2 := User{
		TenantRoles: []UserTenantRole{
			{TenantID: uuid.New()},
			{TenantID: uuid.New()},
		},
	}
	if u2.TenantCount() != 2 {
		t.Errorf("TenantCount() = %d, want 2", u2.TenantCount())
	}
}

func TestUser_NeedsPasswordReset(t *testing.T) {
	u1 := User{MustResetPwd: true}
	if !u1.NeedsPasswordReset() {
		t.Error("Should need password reset")
	}

	u2 := User{MustResetPwd: false}
	if u2.NeedsPasswordReset() {
		t.Error("Should not need password reset")
	}
}

func TestUser_TableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("TableName() = %q, want %q", u.TableName(), "users")
	}
}
