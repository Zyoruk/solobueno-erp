package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/service"
)

func TestToLoginResponse(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	expiresAt := time.Now().Add(time.Hour)

	resp := &service.LoginResponse{
		TokenPair: &domain.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			ExpiresAt:    expiresAt,
		},
		User: &domain.User{
			ID:           userID,
			Email:        "test@example.com",
			FirstName:    "Test",
			LastName:     "User",
			IsActive:     true,
			MustResetPwd: false,
			CreatedAt:    time.Now(),
		},
		TenantID: tenantID,
		Role:     domain.RoleManager,
	}

	result := ToLoginResponse(resp)

	if result.AccessToken != "access-token" {
		t.Error("AccessToken mismatch")
	}
	if result.RefreshToken != "refresh-token" {
		t.Error("RefreshToken mismatch")
	}
	if result.TokenType != "Bearer" {
		t.Error("TokenType mismatch")
	}
	if result.User.ID != userID {
		t.Error("User.ID mismatch")
	}
	if result.User.Email != "test@example.com" {
		t.Error("User.Email mismatch")
	}
	if result.User.Role != "manager" {
		t.Errorf("User.Role = %q, want %q", result.User.Role, "manager")
	}
}

func TestToTokenResponse(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	tp := &domain.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    expiresAt,
	}

	result := ToTokenResponse(tp)

	if result.AccessToken != "new-access-token" {
		t.Error("AccessToken mismatch")
	}
	if result.RefreshToken != "new-refresh-token" {
		t.Error("RefreshToken mismatch")
	}
	if result.TokenType != "Bearer" {
		t.Error("TokenType mismatch")
	}
}

func TestToUserResponse(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	user := &domain.User{
		ID:           userID,
		Email:        "user@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		IsActive:     true,
		MustResetPwd: true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := ToUserResponse(user)

	if result.ID != userID {
		t.Error("ID mismatch")
	}
	if result.Email != "user@example.com" {
		t.Error("Email mismatch")
	}
	if result.FirstName != "John" {
		t.Error("FirstName mismatch")
	}
	if result.LastName != "Doe" {
		t.Error("LastName mismatch")
	}
	if !result.IsActive {
		t.Error("IsActive should be true")
	}
	if !result.MustResetPassword {
		t.Error("MustResetPassword should be true")
	}
}

func TestToTenantOptions(t *testing.T) {
	tenants := []service.TenantInfo{
		{
			ID:   uuid.New(),
			Name: "Restaurant A",
			Slug: "restaurant-a",
			Role: domain.RoleOwner,
		},
		{
			ID:   uuid.New(),
			Name: "Restaurant B",
			Slug: "restaurant-b",
			Role: domain.RoleManager,
		},
	}

	result := ToTenantOptions(tenants)

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}

	if result[0].Name != "Restaurant A" {
		t.Error("First tenant name mismatch")
	}
	if result[0].Slug != "restaurant-a" {
		t.Error("First tenant slug mismatch")
	}

	if result[1].Name != "Restaurant B" {
		t.Error("Second tenant name mismatch")
	}
}

func TestToTenantOptions_Empty(t *testing.T) {
	result := ToTenantOptions([]service.TenantInfo{})

	if len(result) != 0 {
		t.Errorf("len(result) = %d, want 0", len(result))
	}
}
