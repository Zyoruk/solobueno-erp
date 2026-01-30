package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a person with access to the system.
// Users have globally unique emails and can have roles in multiple tenants.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"` // Never serialize
	FirstName    string    `gorm:"size:100;not null" json:"first_name"`
	LastName     string    `gorm:"size:100;not null" json:"last_name"`
	IsActive     bool      `gorm:"default:true;not null;index" json:"is_active"`
	MustResetPwd bool      `gorm:"column:must_reset_pwd;default:false;not null" json:"must_reset_password"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	TenantRoles []UserTenantRole `gorm:"foreignKey:UserID" json:"tenant_roles,omitempty"`
	Sessions    []Session        `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for GORM.
func (User) TableName() string {
	return "users"
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// CanLogin checks if the user is allowed to log in.
func (u *User) CanLogin() bool {
	return u.IsActive
}

// HasTenant checks if the user has access to the specified tenant.
func (u *User) HasTenant(tenantID uuid.UUID) bool {
	for _, tr := range u.TenantRoles {
		if tr.TenantID == tenantID {
			return true
		}
	}
	return false
}

// GetRoleForTenant returns the user's role in the specified tenant.
// Returns empty Role if user doesn't belong to the tenant.
func (u *User) GetRoleForTenant(tenantID uuid.UUID) Role {
	for _, tr := range u.TenantRoles {
		if tr.TenantID == tenantID {
			return tr.Role
		}
	}
	return ""
}

// TenantCount returns the number of tenants the user belongs to.
func (u *User) TenantCount() int {
	return len(u.TenantRoles)
}

// NeedsPasswordReset checks if the user must reset their password.
func (u *User) NeedsPasswordReset() bool {
	return u.MustResetPwd
}
