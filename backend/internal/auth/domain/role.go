// Package domain contains the core domain entities for the auth module.
package domain

import (
	"database/sql/driver"
	"fmt"
)

// Role represents a user's role within a tenant.
type Role string

// Role constants define all available roles in the system.
const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleCashier Role = "cashier"
	RoleWaiter  Role = "waiter"
	RoleKitchen Role = "kitchen"
	RoleViewer  Role = "viewer"
)

// AllRoles returns all valid roles.
func AllRoles() []Role {
	return []Role{
		RoleOwner,
		RoleAdmin,
		RoleManager,
		RoleCashier,
		RoleWaiter,
		RoleKitchen,
		RoleViewer,
	}
}

// IsValid checks if the role is a valid role.
func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleManager, RoleCashier, RoleWaiter, RoleKitchen, RoleViewer:
		return true
	}
	return false
}

// Level returns the numeric level of the role for comparison.
// Higher level = more permissions.
func (r Role) Level() int {
	levels := map[Role]int{
		RoleOwner:   100,
		RoleAdmin:   90,
		RoleManager: 70,
		RoleCashier: 50,
		RoleWaiter:  40,
		RoleKitchen: 30,
		RoleViewer:  10,
	}
	return levels[r]
}

// CanManage returns true if this role can manage users with the other role.
// A role can only manage roles with lower levels.
func (r Role) CanManage(other Role) bool {
	return r.Level() > other.Level()
}

// CanAssign returns true if this role can assign the other role to users.
// Same as CanManage - can only assign roles lower than your own.
func (r Role) CanAssign(other Role) bool {
	return r.CanManage(other)
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// GormDataType implements GORM's custom type interface for migrations.
func (r Role) GormDataType() string {
	return "varchar(20)"
}

// Scan implements sql.Scanner interface for database reads.
func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*r = Role(v)
	case []byte:
		*r = Role(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into Role", value)
	}

	return nil
}

// Value implements driver.Valuer interface for database writes.
func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

// ParseRole parses a string into a Role, returning an error if invalid.
func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return r, nil
}
