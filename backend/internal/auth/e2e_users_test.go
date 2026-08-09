package auth

import (
	"net/http"
	"testing"

	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/handler"
)

// TestE2E_ManagerCreatesAndManagesStaff drives the full staff lifecycle over
// real HTTP: an owner creates a waiter, the new hire logs in with the
// emailed temp password and sets their own, then the owner lists, reads,
// updates, and promotes them.
func TestE2E_ManagerCreatesAndManagesStaff(t *testing.T) {
	env := setupE2E(t)
	tenant := env.seedTenant("Acme Diner", "acme-diner")
	env.seedUser("owner@example.com", "OwnerPass123!", tenant.ID, domain.RoleOwner)

	ownerToken, _, resp := env.login("owner@example.com", "OwnerPass123!")
	resp.Body.Close()
	if ownerToken == "" {
		t.Fatal("owner login did not return an access token")
	}

	// 1. Create a new staff account.
	createResp := env.do(http.MethodPost, "/users", ownerToken, handler.CreateUserRequest{
		Email: "waiter@example.com", FirstName: "New", LastName: "Hire", Role: domain.RoleWaiter,
	})
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	var created handler.CreateUserResponse
	decodeBody(t, createResp, &created)
	if created.MustResetPassword != true || created.LinkedExistingAccount {
		t.Errorf("unexpected create response: %+v", created)
	}

	// 2. The temp password only ever reached the user via email, never the
	// API response - retrieve it from the capturing emailer.
	tempPassword := env.emailer.tempPasswordFor(t, "waiter@example.com")

	// 3. New hire logs in with the temp password.
	staffToken, _, loginResp := env.login("waiter@example.com", tempPassword)
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("staff login status = %d, want %d", loginResp.StatusCode, http.StatusOK)
	}
	loginResp.Body.Close()
	if staffToken == "" {
		t.Fatal("staff login did not return an access token")
	}

	// 4. New hire sets their own password (required before must_reset_password clears).
	changeResp := env.do(http.MethodPost, "/change-password", staffToken, handler.ChangePasswordRequest{
		CurrentPassword: tempPassword, NewPassword: "StaffChosen123!",
	})
	if changeResp.StatusCode != http.StatusOK {
		t.Fatalf("change-password status = %d, want %d", changeResp.StatusCode, http.StatusOK)
	}
	changeResp.Body.Close()

	// 5. Owner lists tenant users and finds the new hire with the right role.
	listResp := env.do(http.MethodGet, "/users", ownerToken, nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, http.StatusOK)
	}
	var list handler.UserListResponse
	decodeBody(t, listResp, &list)
	var staffID string
	for _, u := range list.Data {
		if u.Email == "waiter@example.com" {
			staffID = u.ID.String()
			if u.Role != string(domain.RoleWaiter) {
				t.Errorf("listed role = %q, want %q", u.Role, domain.RoleWaiter)
			}
		}
	}
	if staffID == "" {
		t.Fatalf("new hire not found in tenant user list: %+v", list.Data)
	}

	// 6. Owner reads the user directly.
	getResp := env.do(http.MethodGet, "/users/"+staffID, ownerToken, nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getResp.StatusCode, http.StatusOK)
	}
	var got handler.UserResponse
	decodeBody(t, getResp, &got)
	if got.Email != "waiter@example.com" || got.Role != string(domain.RoleWaiter) {
		t.Errorf("unexpected get response: %+v", got)
	}

	// 7. Owner updates the new hire's name.
	newName := "Updated"
	updateResp := env.do(http.MethodPatch, "/users/"+staffID, ownerToken, handler.UpdateUserRequest{FirstName: &newName})
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, http.StatusOK)
	}
	var updated handler.UserResponse
	decodeBody(t, updateResp, &updated)
	if updated.FirstName != "Updated" {
		t.Errorf("FirstName = %q, want %q", updated.FirstName, "Updated")
	}

	// 8. Owner promotes the new hire to cashier.
	roleResp := env.do(http.MethodPatch, "/users/"+staffID+"/role", ownerToken, handler.UpdateRoleRequest{Role: domain.RoleCashier})
	if roleResp.StatusCode != http.StatusOK {
		t.Fatalf("role update status = %d, want %d", roleResp.StatusCode, http.StatusOK)
	}
	var promoted handler.UserResponse
	decodeBody(t, roleResp, &promoted)
	if promoted.Role != string(domain.RoleCashier) {
		t.Errorf("Role after promotion = %q, want %q", promoted.Role, domain.RoleCashier)
	}

	// 9. The promotion is durable - re-reading shows cashier, not waiter.
	getResp2 := env.do(http.MethodGet, "/users/"+staffID, ownerToken, nil)
	var got2 handler.UserResponse
	decodeBody(t, getResp2, &got2)
	if got2.Role != string(domain.RoleCashier) {
		t.Errorf("Role after re-read = %q, want %q", got2.Role, domain.RoleCashier)
	}
}

// TestE2E_CreateUser_LinksExistingAccountAcrossTenants covers FR-012: an
// email that already has a global account in a different tenant gets a new
// role linked to it instead of a duplicate account/password, and the same
// email colliding within the SAME tenant is still rejected.
func TestE2E_CreateUser_LinksExistingAccountAcrossTenants(t *testing.T) {
	env := setupE2E(t)
	tenantA := env.seedTenant("Tenant A", "tenant-a")
	tenantB := env.seedTenant("Tenant B", "tenant-b")
	env.seedUser("admin-a@example.com", "AdminPass123!", tenantA.ID, domain.RoleOwner)
	env.seedUser("admin-b@example.com", "AdminPass123!", tenantB.ID, domain.RoleOwner)

	adminAToken, _, r1 := env.login("admin-a@example.com", "AdminPass123!")
	r1.Body.Close()
	adminBToken, _, r2 := env.login("admin-b@example.com", "AdminPass123!")
	r2.Body.Close()

	// Admin A creates a shared-email user in tenant A.
	createResp := env.do(http.MethodPost, "/users", adminAToken, handler.CreateUserRequest{
		Email: "shared@example.com", FirstName: "Shared", LastName: "Person", Role: domain.RoleWaiter,
	})
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("initial create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	var created handler.CreateUserResponse
	decodeBody(t, createResp, &created)

	// Admin B creates a user with the SAME email for tenant B - must link,
	// not duplicate.
	linkResp := env.do(http.MethodPost, "/users", adminBToken, handler.CreateUserRequest{
		Email: "shared@example.com", FirstName: "Shared", LastName: "Person", Role: domain.RoleManager,
	})
	if linkResp.StatusCode != http.StatusOK {
		t.Fatalf("linking create status = %d, want %d", linkResp.StatusCode, http.StatusOK)
	}
	var linked handler.CreateUserResponse
	decodeBody(t, linkResp, &linked)
	if !linked.LinkedExistingAccount {
		t.Errorf("expected linked_existing_account=true, got %+v", linked)
	}
	if linked.ID != created.ID {
		t.Errorf("linked user ID = %s, want the original user's ID %s", linked.ID, created.ID)
	}
	if !env.emailer.hasTenantLinkNotification("shared@example.com") {
		t.Error("expected the existing user to be notified of the new tenant link")
	}

	// Admin A tries to add the SAME email to tenant A again - real conflict,
	// must still be rejected.
	dupResp := env.do(http.MethodPost, "/users", adminAToken, handler.CreateUserRequest{
		Email: "shared@example.com", FirstName: "Shared", LastName: "Person", Role: domain.RoleCashier,
	})
	if dupResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("same-tenant duplicate status = %d, want %d", dupResp.StatusCode, http.StatusBadRequest)
	}
	dupResp.Body.Close()

	// The linked user can log in and see both tenants when they select one.
	_, _, loginNoTenant := env.login("shared@example.com", env.emailer.tempPasswordFor(t, "shared@example.com"))
	if loginNoTenant.StatusCode != http.StatusBadRequest {
		t.Errorf("multi-tenant login without tenant_id status = %d, want %d (tenant_required)", loginNoTenant.StatusCode, http.StatusBadRequest)
	}
	loginNoTenant.Body.Close()
}

// TestE2E_AccountLockoutAndUnlock covers FR-011a end to end: 5 failed
// attempts locks the account (423) independent of the correct password,
// and a Manager+ can clear the lockout via POST /users/{id}/unlock.
func TestE2E_AccountLockoutAndUnlock(t *testing.T) {
	env := setupE2E(t)
	tenant := env.seedTenant("Acme Diner", "acme-diner")
	env.seedUser("owner@example.com", "OwnerPass123!", tenant.ID, domain.RoleOwner)
	staff := env.seedUser("staff@example.com", "StaffPass123!", tenant.ID, domain.RoleWaiter)

	ownerToken, _, ownerLogin := env.login("owner@example.com", "OwnerPass123!")
	ownerLogin.Body.Close()

	for i := 0; i < 5; i++ {
		resp := env.do(http.MethodPost, "/login", "", map[string]string{"email": "staff@example.com", "password": "WrongPassword!"})
		resp.Body.Close()
	}

	lockedResp := env.do(http.MethodPost, "/login", "", map[string]string{"email": "staff@example.com", "password": "StaffPass123!"})
	if lockedResp.StatusCode != http.StatusLocked {
		t.Fatalf("locked login status = %d, want %d", lockedResp.StatusCode, http.StatusLocked)
	}
	var lockedBody handler.ErrorResponse
	decodeBody(t, lockedResp, &lockedBody)
	if lockedBody.Error.Code != "account_locked" || lockedBody.Error.LockedUntil == nil {
		t.Errorf("unexpected lockout error body: %+v", lockedBody.Error)
	}

	unlockResp := env.do(http.MethodPost, "/users/"+staff.ID.String()+"/unlock", ownerToken, nil)
	if unlockResp.StatusCode != http.StatusOK {
		t.Fatalf("unlock status = %d, want %d", unlockResp.StatusCode, http.StatusOK)
	}
	unlockResp.Body.Close()

	_, _, afterUnlock := env.login("staff@example.com", "StaffPass123!")
	if afterUnlock.StatusCode != http.StatusOK {
		t.Fatalf("post-unlock login status = %d, want %d", afterUnlock.StatusCode, http.StatusOK)
	}
	afterUnlock.Body.Close()
}
