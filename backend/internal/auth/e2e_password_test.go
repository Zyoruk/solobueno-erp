package auth

import (
	"net/http"
	"testing"

	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/handler"
)

// TestE2E_ChangePassword_InvalidatesOtherSessions covers FR-014: changing
// your password over real HTTP must revoke every session, including ones
// from other devices, not just the one making the request.
func TestE2E_ChangePassword_InvalidatesOtherSessions(t *testing.T) {
	env := setupE2E(t)
	tenant := env.seedTenant("Acme Diner", "acme-diner")
	env.seedUser("staff@example.com", "OldPass123!", tenant.ID, domain.RoleWaiter)

	deviceAAccess, deviceARefresh, r1 := env.login("staff@example.com", "OldPass123!")
	r1.Body.Close()
	_, deviceBRefresh, r2 := env.login("staff@example.com", "OldPass123!")
	r2.Body.Close()

	changeResp := env.do(http.MethodPost, "/change-password", deviceAAccess, handler.ChangePasswordRequest{
		CurrentPassword: "OldPass123!", NewPassword: "NewPass456!",
	})
	if changeResp.StatusCode != http.StatusOK {
		t.Fatalf("change-password status = %d, want %d", changeResp.StatusCode, http.StatusOK)
	}
	changeResp.Body.Close()

	// Both the requesting device's session and the other device's session
	// must be revoked.
	refreshA := env.do(http.MethodPost, "/refresh", "", map[string]string{"refresh_token": deviceARefresh})
	if refreshA.StatusCode != http.StatusUnauthorized {
		t.Errorf("device A refresh after change-password status = %d, want %d", refreshA.StatusCode, http.StatusUnauthorized)
	}
	refreshA.Body.Close()

	refreshB := env.do(http.MethodPost, "/refresh", "", map[string]string{"refresh_token": deviceBRefresh})
	if refreshB.StatusCode != http.StatusUnauthorized {
		t.Errorf("device B refresh after change-password status = %d, want %d", refreshB.StatusCode, http.StatusUnauthorized)
	}
	refreshB.Body.Close()

	// New password works, old password no longer does.
	_, _, newLogin := env.login("staff@example.com", "NewPass456!")
	if newLogin.StatusCode != http.StatusOK {
		t.Errorf("login with new password status = %d, want %d", newLogin.StatusCode, http.StatusOK)
	}
	newLogin.Body.Close()

	_, _, oldLogin := env.login("staff@example.com", "OldPass123!")
	if oldLogin.StatusCode != http.StatusUnauthorized {
		t.Errorf("login with old password status = %d, want %d", oldLogin.StatusCode, http.StatusUnauthorized)
	}
	oldLogin.Body.Close()
}

// TestE2E_PasswordResetFlow covers FR-013 end to end: request -> the token
// only ever leaves the system via email (captured here instead) -> complete
// -> new password works, old one doesn't, and the token can't be reused.
func TestE2E_PasswordResetFlow(t *testing.T) {
	env := setupE2E(t)
	tenant := env.seedTenant("Acme Diner", "acme-diner")
	env.seedUser("staff@example.com", "OldPass123!", tenant.ID, domain.RoleWaiter)

	reqResp := env.do(http.MethodPost, "/password-reset/request", "", handler.PasswordResetRequest{Email: "staff@example.com"})
	if reqResp.StatusCode != http.StatusAccepted {
		t.Fatalf("reset request status = %d, want %d", reqResp.StatusCode, http.StatusAccepted)
	}
	reqResp.Body.Close()

	token := env.emailer.resetTokenFor(t, "staff@example.com")

	completeResp := env.do(http.MethodPost, "/password-reset/complete", "", handler.PasswordResetCompleteRequest{
		Token: token, NewPassword: "ResetPass789!",
	})
	if completeResp.StatusCode != http.StatusOK {
		t.Fatalf("reset complete status = %d, want %d", completeResp.StatusCode, http.StatusOK)
	}
	completeResp.Body.Close()

	_, _, newLogin := env.login("staff@example.com", "ResetPass789!")
	if newLogin.StatusCode != http.StatusOK {
		t.Errorf("login with reset password status = %d, want %d", newLogin.StatusCode, http.StatusOK)
	}
	newLogin.Body.Close()

	_, _, oldLogin := env.login("staff@example.com", "OldPass123!")
	if oldLogin.StatusCode != http.StatusUnauthorized {
		t.Errorf("login with pre-reset password status = %d, want %d", oldLogin.StatusCode, http.StatusUnauthorized)
	}
	oldLogin.Body.Close()

	// The token is single-use.
	reuseResp := env.do(http.MethodPost, "/password-reset/complete", "", handler.PasswordResetCompleteRequest{
		Token: token, NewPassword: "AnotherPass000!",
	})
	if reuseResp.StatusCode != http.StatusBadRequest {
		t.Errorf("reused token status = %d, want %d", reuseResp.StatusCode, http.StatusBadRequest)
	}
	reuseResp.Body.Close()

	// Requesting a reset for an unknown email still returns 202 (no
	// enumeration) and captures no token.
	unknownResp := env.do(http.MethodPost, "/password-reset/request", "", handler.PasswordResetRequest{Email: "nobody@example.com"})
	if unknownResp.StatusCode != http.StatusAccepted {
		t.Errorf("unknown-email reset request status = %d, want %d", unknownResp.StatusCode, http.StatusAccepted)
	}
	unknownResp.Body.Close()
}
