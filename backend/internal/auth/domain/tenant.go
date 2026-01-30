package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a restaurant/business entity.
type Tenant struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	IsActive  bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	UserRoles []UserTenantRole `gorm:"foreignKey:TenantID" json:"-"`
}

// TableName specifies the table name for GORM.
func (Tenant) TableName() string {
	return "tenants"
}

// IsOperational checks if the tenant is active and can be used.
func (t *Tenant) IsOperational() bool {
	return t.IsActive
}
