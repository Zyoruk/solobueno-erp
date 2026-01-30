package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
	"gorm.io/gorm"
)

// SeedData contains the seed data for testing.
type SeedData struct {
	Tenant      *domain.Tenant
	OwnerUser   *domain.User
	ManagerUser *domain.User
	WaiterUser  *domain.User
	Passwords   map[string]string // email -> plain password
}

// Seed creates test data for development and testing.
func Seed(ctx context.Context, db *gorm.DB) (*SeedData, error) {
	passwordSvc := service.NewPasswordService()

	// Create a test tenant
	tenant := &domain.Tenant{
		ID:       uuid.New(),
		Name:     "Test Restaurant",
		Slug:     "test-restaurant",
		IsActive: true,
	}
	if err := db.WithContext(ctx).Create(tenant).Error; err != nil {
		return nil, err
	}

	passwords := make(map[string]string)

	// Create owner user
	ownerPassword := "Owner123!"
	ownerHash, _ := passwordSvc.Hash(ownerPassword)
	ownerUser := &domain.User{
		ID:           uuid.New(),
		Email:        "owner@test.com",
		PasswordHash: ownerHash,
		FirstName:    "Test",
		LastName:     "Owner",
		IsActive:     true,
		MustResetPwd: false,
	}
	if err := db.WithContext(ctx).Create(ownerUser).Error; err != nil {
		return nil, err
	}
	passwords["owner@test.com"] = ownerPassword

	// Assign owner role
	ownerRole := &domain.UserTenantRole{
		ID:       uuid.New(),
		UserID:   ownerUser.ID,
		TenantID: tenant.ID,
		Role:     domain.RoleOwner,
	}
	if err := db.WithContext(ctx).Create(ownerRole).Error; err != nil {
		return nil, err
	}

	// Create manager user
	managerPassword := "Manager123!"
	managerHash, _ := passwordSvc.Hash(managerPassword)
	managerUser := &domain.User{
		ID:           uuid.New(),
		Email:        "manager@test.com",
		PasswordHash: managerHash,
		FirstName:    "Test",
		LastName:     "Manager",
		IsActive:     true,
		MustResetPwd: false,
	}
	if err := db.WithContext(ctx).Create(managerUser).Error; err != nil {
		return nil, err
	}
	passwords["manager@test.com"] = managerPassword

	// Assign manager role
	managerRole := &domain.UserTenantRole{
		ID:       uuid.New(),
		UserID:   managerUser.ID,
		TenantID: tenant.ID,
		Role:     domain.RoleManager,
	}
	if err := db.WithContext(ctx).Create(managerRole).Error; err != nil {
		return nil, err
	}

	// Create waiter user
	waiterPassword := "Waiter123!"
	waiterHash, _ := passwordSvc.Hash(waiterPassword)
	waiterUser := &domain.User{
		ID:           uuid.New(),
		Email:        "waiter@test.com",
		PasswordHash: waiterHash,
		FirstName:    "Test",
		LastName:     "Waiter",
		IsActive:     true,
		MustResetPwd: false,
	}
	if err := db.WithContext(ctx).Create(waiterUser).Error; err != nil {
		return nil, err
	}
	passwords["waiter@test.com"] = waiterPassword

	// Assign waiter role
	waiterRole := &domain.UserTenantRole{
		ID:       uuid.New(),
		UserID:   waiterUser.ID,
		TenantID: tenant.ID,
		Role:     domain.RoleWaiter,
	}
	if err := db.WithContext(ctx).Create(waiterRole).Error; err != nil {
		return nil, err
	}

	return &SeedData{
		Tenant:      tenant,
		OwnerUser:   ownerUser,
		ManagerUser: managerUser,
		WaiterUser:  waiterUser,
		Passwords:   passwords,
	}, nil
}
