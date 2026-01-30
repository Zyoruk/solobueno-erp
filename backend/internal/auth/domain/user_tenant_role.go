package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserTenantRole represents the junction between users and tenants,
// defining what role a user has within a specific tenant.
type UserTenantRole struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_tenant" json:"user_id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_tenant" json:"tenant_id"`
	Role      Role      `gorm:"size:20;not null;index" json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	User   User   `gorm:"foreignKey:UserID" json:"-"`
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

// TableName specifies the table name for GORM.
func (UserTenantRole) TableName() string {
	return "user_tenant_roles"
}
