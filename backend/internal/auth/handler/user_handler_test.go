package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
	"github.com/solobueno/erp/internal/auth/repository/mock"
	"github.com/solobueno/erp/internal/auth/service"
)

func setupUserHandler(t *testing.T) (*UserHandler, *mock.MockUserRepository, *mock.MockUserTenantRoleRepository) {
	t.Helper()

	userRepo := mock.NewMockUserRepository()
	roleRepo := mock.NewMockUserTenantRoleRepository()

	userSvc := service.NewUserService(service.UserServiceConfig{
		UserRepo:      userRepo,
		RoleRepo:      roleRepo,
		SessionRepo:   mock.NewMockSessionRepository(),
		EventRepo:     mock.NewMockAuthEventRepository(),
		PasswordReset: mock.NewMockPasswordResetRepository(),
	})

	return NewUserHandler(userSvc), userRepo, roleRepo
}

// authedContext builds a request context as RequireAuth would populate it.
func authedContext(userID, tenantID uuid.UUID, role domain.Role) context.Context {
	ctx := context.WithValue(context.Background(), UserIDContextKey, userID)
	ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)
	ctx = context.WithValue(ctx, RoleContextKey, role)
	return ctx
}

// withChiURLParam attaches a chi route param the way the router would when
// matching "/users/{id}".
func withChiURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestUserHandler_Create_Success(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	tenantID := uuid.New()
	body, _ := json.Marshal(CreateUserRequest{Email: "new@example.com", FirstName: "New", LastName: "Hire", Role: domain.RoleWaiter})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), tenantID, domain.RoleManager))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var resp CreateUserResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Email != "new@example.com" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if strings.Contains(w.Body.String(), "temporary_password") {
		t.Errorf("response must never include the temporary password: %s", w.Body.String())
	}
}

func TestUserHandler_Create_Unauthorized(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	body, _ := json.Marshal(CreateUserRequest{Email: "x@example.com", FirstName: "X", LastName: "Y", Role: domain.RoleWaiter})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_Create_MissingFields(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	body, _ := json.Marshal(CreateUserRequest{Role: domain.RoleWaiter})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Create_InvalidRole(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	body, _ := json.Marshal(CreateUserRequest{Email: "x@example.com", FirstName: "X", LastName: "Y", Role: domain.Role("not-a-role")})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Create_InvalidBody(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader([]byte("invalid json"))).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Create_CannotAssignRole(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	// A waiter cannot create another waiter (must outrank the assigned role).
	body, _ := json.Marshal(CreateUserRequest{Email: "x@example.com", FirstName: "X", LastName: "Y", Role: domain.RoleWaiter})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleWaiter))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d, body=%s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestUserHandler_Create_EmailExistsInSameTenant(t *testing.T) {
	h, userRepo, _ := setupUserHandler(t)
	tenantID := uuid.New()
	existingID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID: existingID, Email: "dup@example.com",
		TenantRoles: []domain.UserTenantRole{{UserID: existingID, TenantID: tenantID, Role: domain.RoleWaiter}},
	})

	body, _ := json.Marshal(CreateUserRequest{Email: "dup@example.com", FirstName: "X", LastName: "Y", Role: domain.RoleWaiter})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), tenantID, domain.RoleManager))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Create_LinksExistingUserToNewTenant(t *testing.T) {
	h, userRepo, _ := setupUserHandler(t)
	existingID := uuid.New()
	otherTenantID := uuid.New()
	userRepo.AddUser(&domain.User{
		ID: existingID, Email: "existing@example.com",
		TenantRoles: []domain.UserTenantRole{{UserID: existingID, TenantID: otherTenantID, Role: domain.RoleWaiter}},
	})

	newTenantID := uuid.New()
	body, _ := json.Marshal(CreateUserRequest{Email: "existing@example.com", FirstName: "X", LastName: "Y", Role: domain.RoleManager})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), newTenantID, domain.RoleAdmin))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp CreateUserResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.LinkedExistingAccount {
		t.Errorf("expected linked_existing_account=true, got %+v", resp)
	}
	if strings.Contains(w.Body.String(), "temporary_password") {
		t.Errorf("linked response must never include a temporary password: %s", w.Body.String())
	}
}

func TestUserHandler_List(t *testing.T) {
	h, userRepo, roleRepo := setupUserHandler(t)

	tenantID := uuid.New()
	user := &domain.User{ID: uuid.New(), Email: "listed@example.com", IsActive: true}
	userRepo.AddUser(user)
	roleRepo.AddRole(&domain.UserTenantRole{ID: uuid.New(), UserID: user.ID, TenantID: tenantID, Role: domain.RoleWaiter})

	req := httptest.NewRequest("GET", "/users", nil).WithContext(authedContext(uuid.New(), tenantID, domain.RoleManager))
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp UserListResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Pagination.Page != 1 || resp.Pagination.Limit != 20 {
		t.Errorf("unexpected pagination defaults: %+v", resp.Pagination)
	}
}

func TestUserHandler_List_Unauthorized(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_Get_Success(t *testing.T) {
	h, userRepo, _ := setupUserHandler(t)

	user := &domain.User{ID: uuid.New(), Email: "found@example.com"}
	userRepo.AddUser(user)

	req := httptest.NewRequest("GET", "/users/"+user.ID.String(), nil)
	req = withChiURLParam(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUserHandler_Get_InvalidID(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("GET", "/users/not-a-uuid", nil)
	req = withChiURLParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	id := uuid.New()
	req := httptest.NewRequest("GET", "/users/"+id.String(), nil)
	req = withChiURLParam(req, "id", id.String())
	w := httptest.NewRecorder()

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUserHandler_Update_Success(t *testing.T) {
	h, userRepo, _ := setupUserHandler(t)

	user := &domain.User{ID: uuid.New(), Email: "update@example.com", FirstName: "Before"}
	userRepo.AddUser(user)

	newName := "After"
	body, _ := json.Marshal(UpdateUserRequest{FirstName: &newName})
	req := httptest.NewRequest("PATCH", "/users/"+user.ID.String(), bytes.NewReader(body)).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUserHandler_Update_Unauthorized(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("PATCH", "/users/"+uuid.New().String(), bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("PATCH", "/users/not-a-uuid", bytes.NewReader([]byte("{}"))).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Update_InvalidBody(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("PATCH", "/users/"+uuid.New().String(), bytes.NewReader([]byte("invalid json"))).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	id := uuid.New()
	req := httptest.NewRequest("PATCH", "/users/"+id.String(), bytes.NewReader([]byte("{}"))).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", id.String())
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUserHandler_Unlock_Success(t *testing.T) {
	h, userRepo, _ := setupUserHandler(t)

	tenantID := uuid.New()
	user := &domain.User{
		ID: uuid.New(), Email: "locked@example.com", FailedLoginCount: 5,
		TenantRoles: []domain.UserTenantRole{{TenantID: tenantID, Role: domain.RoleWaiter}},
	}
	userRepo.AddUser(user)

	req := httptest.NewRequest("POST", "/users/"+user.ID.String()+"/unlock", nil).WithContext(authedContext(uuid.New(), tenantID, domain.RoleManager))
	req = withChiURLParam(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.Unlock(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUserHandler_Unlock_Unauthorized(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("POST", "/users/"+uuid.New().String()+"/unlock", nil)
	w := httptest.NewRecorder()

	h.Unlock(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_Unlock_InvalidID(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("POST", "/users/not-a-uuid/unlock", nil).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Unlock(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_Unlock_NotFound(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	id := uuid.New()
	req := httptest.NewRequest("POST", "/users/"+id.String()+"/unlock", nil).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", id.String())
	w := httptest.NewRecorder()

	h.Unlock(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUserHandler_UpdateRole_Success(t *testing.T) {
	h, userRepo, roleRepo := setupUserHandler(t)

	tenantID := uuid.New()
	user := &domain.User{ID: uuid.New(), Email: "role@example.com"}
	userRepo.AddUser(user)
	roleRepo.AddRole(&domain.UserTenantRole{ID: uuid.New(), UserID: user.ID, TenantID: tenantID, Role: domain.RoleWaiter})

	body, _ := json.Marshal(UpdateRoleRequest{Role: domain.RoleCashier})
	req := httptest.NewRequest("PATCH", "/users/"+user.ID.String()+"/role", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), tenantID, domain.RoleManager))
	req = withChiURLParam(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.UpdateRole(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUserHandler_UpdateRole_Unauthorized(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("PATCH", "/users/"+uuid.New().String()+"/role", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	h.UpdateRole(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestUserHandler_UpdateRole_InvalidID(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("PATCH", "/users/not-a-uuid/role", bytes.NewReader([]byte("{}"))).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.UpdateRole(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateRole_InvalidBody(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	req := httptest.NewRequest("PATCH", "/users/"+uuid.New().String()+"/role", bytes.NewReader([]byte("invalid json"))).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	h.UpdateRole(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateRole_InvalidRole(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	body, _ := json.Marshal(UpdateRoleRequest{Role: domain.Role("bogus")})
	req := httptest.NewRequest("PATCH", "/users/"+uuid.New().String()+"/role", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	h.UpdateRole(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_UpdateRole_NotInTenant(t *testing.T) {
	h, userRepo, _ := setupUserHandler(t)

	user := &domain.User{ID: uuid.New(), Email: "notenant@example.com"}
	userRepo.AddUser(user)

	body, _ := json.Marshal(UpdateRoleRequest{Role: domain.RoleCashier})
	req := httptest.NewRequest("PATCH", "/users/"+user.ID.String()+"/role", bytes.NewReader(body)).WithContext(authedContext(uuid.New(), uuid.New(), domain.RoleManager))
	req = withChiURLParam(req, "id", user.ID.String())
	w := httptest.NewRecorder()

	h.UpdateRole(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d, body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}
