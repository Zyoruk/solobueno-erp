package domain

import (
	"testing"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		role  Role
		valid bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleManager, true},
		{RoleCashier, true},
		{RoleWaiter, true},
		{RoleKitchen, true},
		{RoleViewer, true},
		{Role("invalid"), false},
		{Role(""), false},
		{Role("superadmin"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.valid {
				t.Errorf("Role(%q).IsValid() = %v, want %v", tt.role, got, tt.valid)
			}
		})
	}
}

func TestRole_Level(t *testing.T) {
	tests := []struct {
		role  Role
		level int
	}{
		{RoleOwner, 100},
		{RoleAdmin, 90},
		{RoleManager, 70},
		{RoleCashier, 50},
		{RoleWaiter, 40},
		{RoleKitchen, 30},
		{RoleViewer, 10},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.Level(); got != tt.level {
				t.Errorf("Role(%q).Level() = %v, want %v", tt.role, got, tt.level)
			}
		})
	}
}

func TestRole_CanManage(t *testing.T) {
	tests := []struct {
		role     Role
		other    Role
		expected bool
	}{
		// Owner can manage everyone
		{RoleOwner, RoleAdmin, true},
		{RoleOwner, RoleManager, true},
		{RoleOwner, RoleViewer, true},
		{RoleOwner, RoleOwner, false}, // Cannot manage same level

		// Admin can manage below
		{RoleAdmin, RoleManager, true},
		{RoleAdmin, RoleOwner, false},
		{RoleAdmin, RoleAdmin, false},

		// Manager can manage staff
		{RoleManager, RoleCashier, true},
		{RoleManager, RoleWaiter, true},
		{RoleManager, RoleKitchen, true},
		{RoleManager, RoleViewer, true},
		{RoleManager, RoleAdmin, false},
		{RoleManager, RoleManager, false},

		// Lower roles cannot manage higher
		{RoleWaiter, RoleManager, false},
		{RoleViewer, RoleWaiter, false},
	}

	for _, tt := range tests {
		name := string(tt.role) + "_manages_" + string(tt.other)
		t.Run(name, func(t *testing.T) {
			if got := tt.role.CanManage(tt.other); got != tt.expected {
				t.Errorf("Role(%q).CanManage(%q) = %v, want %v", tt.role, tt.other, got, tt.expected)
			}
		})
	}
}

func TestRole_CanAssign(t *testing.T) {
	// CanAssign has same logic as CanManage
	if RoleManager.CanAssign(RoleWaiter) != RoleManager.CanManage(RoleWaiter) {
		t.Error("CanAssign should have same behavior as CanManage")
	}
}

func TestRole_String(t *testing.T) {
	if RoleManager.String() != "manager" {
		t.Errorf("RoleManager.String() = %q, want %q", RoleManager.String(), "manager")
	}
}

func TestRole_GormDataType(t *testing.T) {
	if RoleManager.GormDataType() != "varchar(20)" {
		t.Errorf("GormDataType() = %q, want %q", RoleManager.GormDataType(), "varchar(20)")
	}
}

func TestRole_ScanAndValue(t *testing.T) {
	var r Role

	// Test Scan with string
	if err := r.Scan("manager"); err != nil {
		t.Errorf("Scan(string) error: %v", err)
	}
	if r != RoleManager {
		t.Errorf("After Scan, role = %q, want %q", r, RoleManager)
	}

	// Test Scan with []byte
	if err := r.Scan([]byte("admin")); err != nil {
		t.Errorf("Scan([]byte) error: %v", err)
	}
	if r != RoleAdmin {
		t.Errorf("After Scan, role = %q, want %q", r, RoleAdmin)
	}

	// Test Scan with nil
	if err := r.Scan(nil); err != nil {
		t.Errorf("Scan(nil) error: %v", err)
	}

	// Test Scan with invalid type
	if err := r.Scan(123); err == nil {
		t.Error("Scan(int) should return error")
	}

	// Test Value
	r = RoleOwner
	val, err := r.Value()
	if err != nil {
		t.Errorf("Value() error: %v", err)
	}
	if val != "owner" {
		t.Errorf("Value() = %v, want %q", val, "owner")
	}
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		input   string
		want    Role
		wantErr bool
	}{
		{"owner", RoleOwner, false},
		{"admin", RoleAdmin, false},
		{"manager", RoleManager, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRole(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRole(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAllRoles(t *testing.T) {
	roles := AllRoles()
	if len(roles) != 7 {
		t.Errorf("AllRoles() returned %d roles, want 7", len(roles))
	}

	// Verify all roles are valid
	for _, r := range roles {
		if !r.IsValid() {
			t.Errorf("AllRoles() contains invalid role: %q", r)
		}
	}
}
