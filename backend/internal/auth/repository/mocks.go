package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solobueno/erp/internal/auth/domain"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User

	// Hooks for custom behavior
	FindByIDFunc               func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmailFunc            func(ctx context.Context, email string) (*domain.User, error)
	FindByEmailWithTenantsFunc func(ctx context.Context, email string) (*domain.User, error)
	FindByIDWithTenantsFunc    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	CreateFunc                 func(ctx context.Context, user *domain.User) error
	UpdateFunc                 func(ctx context.Context, user *domain.User) error
	ListByTenantFunc           func(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.User, int64, error)
	ExistsByEmailFunc          func(ctx context.Context, email string) (bool, error)
}

// NewMockUserRepository creates a new MockUserRepository.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockUserRepository) FindByEmailWithTenants(ctx context.Context, email string) (*domain.User, error) {
	if m.FindByEmailWithTenantsFunc != nil {
		return m.FindByEmailWithTenantsFunc(ctx, email)
	}
	return m.FindByEmail(ctx, email)
}

func (m *MockUserRepository) FindByIDWithTenants(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.FindByIDWithTenantsFunc != nil {
		return m.FindByIDWithTenantsFunc(ctx, id)
	}
	return m.FindByID(ctx, id)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.User, int64, error) {
	if m.ListByTenantFunc != nil {
		return m.ListByTenantFunc(ctx, tenantID, offset, limit)
	}
	return []*domain.User{}, 0, nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.ExistsByEmailFunc != nil {
		return m.ExistsByEmailFunc(ctx, email)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// AddUser adds a user to the mock repository.
func (m *MockUserRepository) AddUser(user *domain.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
}

var _ UserRepository = (*MockUserRepository)(nil)

// MockSessionRepository is a mock implementation of SessionRepository.
type MockSessionRepository struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*domain.Session

	CreateFunc                  func(ctx context.Context, session *domain.Session) error
	FindByTokenFunc             func(ctx context.Context, tokenHash string) (*domain.Session, error)
	FindByIDFunc                func(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	RevokeFunc                  func(ctx context.Context, id uuid.UUID) error
	RevokeByTokenFunc           func(ctx context.Context, tokenHash string) error
	RevokeAllForUserFunc        func(ctx context.Context, userID uuid.UUID) error
	RevokeAllForUserInTenantFunc func(ctx context.Context, userID, tenantID uuid.UUID) error
	DeleteExpiredFunc           func(ctx context.Context) (int64, error)
	CountActiveForUserFunc      func(ctx context.Context, userID uuid.UUID) (int64, error)
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[uuid.UUID]*domain.Session),
	}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, session)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionRepository) FindByToken(ctx context.Context, tokenHash string) (*domain.Session, error) {
	if m.FindByTokenFunc != nil {
		return m.FindByTokenFunc(ctx, tokenHash)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.sessions {
		if s.RefreshToken == tokenHash {
			return s, nil
		}
	}
	return nil, domain.ErrSessionNotFound
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (m *MockSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(ctx, id)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.Revoke()
	}
	return nil
}

func (m *MockSessionRepository) RevokeByToken(ctx context.Context, tokenHash string) error {
	if m.RevokeByTokenFunc != nil {
		return m.RevokeByTokenFunc(ctx, tokenHash)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.RefreshToken == tokenHash {
			s.Revoke()
			return nil
		}
	}
	return domain.ErrSessionNotFound
}

func (m *MockSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if m.RevokeAllForUserFunc != nil {
		return m.RevokeAllForUserFunc(ctx, userID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.UserID == userID {
			s.Revoke()
		}
	}
	return nil
}

func (m *MockSessionRepository) RevokeAllForUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	if m.RevokeAllForUserInTenantFunc != nil {
		return m.RevokeAllForUserInTenantFunc(ctx, userID, tenantID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.UserID == userID && s.TenantID == tenantID {
			s.Revoke()
		}
	}
	return nil
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc(ctx)
	}
	return 0, nil
}

func (m *MockSessionRepository) CountActiveForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.CountActiveForUserFunc != nil {
		return m.CountActiveForUserFunc(ctx, userID)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	now := time.Now()
	for _, s := range m.sessions {
		if s.UserID == userID && s.RevokedAt == nil && s.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

var _ SessionRepository = (*MockSessionRepository)(nil)

// MockAuthEventRepository is a mock implementation of AuthEventRepository.
type MockAuthEventRepository struct {
	mu     sync.RWMutex
	events []*domain.AuthEvent

	CreateFunc          func(ctx context.Context, event *domain.AuthEvent) error
	FindByUserFunc      func(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error)
	CountRecentByIPFunc func(ctx context.Context, ipAddress string, eventType domain.AuthEventType, since time.Time) (int64, error)
}

func NewMockAuthEventRepository() *MockAuthEventRepository {
	return &MockAuthEventRepository{
		events: make([]*domain.AuthEvent, 0),
	}
}

func (m *MockAuthEventRepository) Create(ctx context.Context, event *domain.AuthEvent) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *MockAuthEventRepository) FindByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error) {
	if m.FindByUserFunc != nil {
		return m.FindByUserFunc(ctx, userID, offset, limit)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.AuthEvent
	for _, e := range m.events {
		if e.UserID != nil && *e.UserID == userID {
			result = append(result, e)
		}
	}
	// Apply pagination
	start := offset
	if start > len(result) {
		start = len(result)
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], int64(len(result)), nil
}

func (m *MockAuthEventRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]*domain.AuthEvent, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.AuthEvent
	for _, e := range m.events {
		if e.TenantID != nil && *e.TenantID == tenantID {
			result = append(result, e)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockAuthEventRepository) FindByType(ctx context.Context, eventType domain.AuthEventType, offset, limit int) ([]*domain.AuthEvent, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.AuthEvent
	for _, e := range m.events {
		if e.EventType == eventType {
			result = append(result, e)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockAuthEventRepository) FindByUserAndType(ctx context.Context, userID uuid.UUID, eventType domain.AuthEventType, since time.Time) ([]*domain.AuthEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.AuthEvent
	for _, e := range m.events {
		if e.UserID != nil && *e.UserID == userID && e.EventType == eventType && e.CreatedAt.After(since) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MockAuthEventRepository) CountRecentByIP(ctx context.Context, ipAddress string, eventType domain.AuthEventType, since time.Time) (int64, error) {
	if m.CountRecentByIPFunc != nil {
		return m.CountRecentByIPFunc(ctx, ipAddress, eventType, since)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, e := range m.events {
		if e.IPAddress == ipAddress && e.EventType == eventType && e.CreatedAt.After(since) {
			count++
		}
	}
	return count, nil
}

func (m *MockAuthEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var kept []*domain.AuthEvent
	var deleted int64
	for _, e := range m.events {
		if e.CreatedAt.After(before) {
			kept = append(kept, e)
		} else {
			deleted++
		}
	}
	m.events = kept
	return deleted, nil
}

// GetEvents returns all recorded events.
func (m *MockAuthEventRepository) GetEvents() []*domain.AuthEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.events
}

var _ AuthEventRepository = (*MockAuthEventRepository)(nil)

// MockTenantRepository is a mock implementation of TenantRepository.
type MockTenantRepository struct {
	mu      sync.RWMutex
	tenants map[uuid.UUID]*domain.Tenant

	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{
		tenants: make(map[uuid.UUID]*domain.Tenant),
	}
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTenantNotFound
}

func (m *MockTenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, domain.ErrTenantNotFound
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockTenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tenants {
		if t.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

// AddTenant adds a tenant to the mock repository.
func (m *MockTenantRepository) AddTenant(tenant *domain.Tenant) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tenants[tenant.ID] = tenant
}

var _ TenantRepository = (*MockTenantRepository)(nil)

// MockUserTenantRoleRepository is a mock implementation of UserTenantRoleRepository.
type MockUserTenantRoleRepository struct {
	mu    sync.RWMutex
	roles map[uuid.UUID]*domain.UserTenantRole

	FindByUserAndTenantFunc func(ctx context.Context, userID, tenantID uuid.UUID) (*domain.UserTenantRole, error)
	CreateFunc              func(ctx context.Context, role *domain.UserTenantRole) error
	UpdateFunc              func(ctx context.Context, role *domain.UserTenantRole) error
}

func NewMockUserTenantRoleRepository() *MockUserTenantRoleRepository {
	return &MockUserTenantRoleRepository{
		roles: make(map[uuid.UUID]*domain.UserTenantRole),
	}
}

func (m *MockUserTenantRoleRepository) FindByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.UserTenantRole, error) {
	if m.FindByUserAndTenantFunc != nil {
		return m.FindByUserAndTenantFunc(ctx, userID, tenantID)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.roles {
		if r.UserID == userID && r.TenantID == tenantID {
			return r, nil
		}
	}
	return nil, domain.ErrUserNotInTenant
}

func (m *MockUserTenantRoleRepository) Create(ctx context.Context, role *domain.UserTenantRole) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, role)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	m.roles[role.ID] = role
	return nil
}

func (m *MockUserTenantRoleRepository) Update(ctx context.Context, role *domain.UserTenantRole) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, role)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *MockUserTenantRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, id)
	return nil
}

func (m *MockUserTenantRoleRepository) DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, r := range m.roles {
		if r.UserID == userID && r.TenantID == tenantID {
			delete(m.roles, id)
			return nil
		}
	}
	return domain.ErrUserNotInTenant
}

func (m *MockUserTenantRoleRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserTenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.UserTenantRole
	for _, r := range m.roles {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockUserTenantRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.UserTenantRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.UserTenantRole
	for _, r := range m.roles {
		if r.TenantID == tenantID {
			result = append(result, r)
		}
	}
	return result, nil
}

// AddRole adds a role assignment to the mock repository.
func (m *MockUserTenantRoleRepository) AddRole(role *domain.UserTenantRole) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
}

var _ UserTenantRoleRepository = (*MockUserTenantRoleRepository)(nil)

// MockPasswordResetRepository is a mock implementation of PasswordResetRepository.
type MockPasswordResetRepository struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]*domain.PasswordResetToken

	FindByTokenFunc func(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)
}

func NewMockPasswordResetRepository() *MockPasswordResetRepository {
	return &MockPasswordResetRepository{
		tokens: make(map[uuid.UUID]*domain.PasswordResetToken),
	}
}

func (m *MockPasswordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	m.tokens[token.ID] = token
	return nil
}

func (m *MockPasswordResetRepository) FindByToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	if m.FindByTokenFunc != nil {
		return m.FindByTokenFunc(ctx, tokenHash)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tokens {
		if t.TokenHash == tokenHash {
			return t, nil
		}
	}
	return nil, domain.ErrPasswordResetInvalid
}

func (m *MockPasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tokens[id]; ok {
		t.MarkUsed()
	}
	return nil
}

func (m *MockPasswordResetRepository) DeleteExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var deleted int64
	now := time.Now()
	for id, t := range m.tokens {
		if t.ExpiresAt.Before(now) {
			delete(m.tokens, id)
			deleted++
		}
	}
	return deleted, nil
}

func (m *MockPasswordResetRepository) DeleteForUser(ctx context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, t := range m.tokens {
		if t.UserID == userID {
			delete(m.tokens, id)
		}
	}
	return nil
}

func (m *MockPasswordResetRepository) CountRecentForUser(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, t := range m.tokens {
		if t.UserID == userID && t.CreatedAt.After(since) {
			count++
		}
	}
	return count, nil
}

// AddToken adds a token to the mock repository.
func (m *MockPasswordResetRepository) AddToken(token *domain.PasswordResetToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[token.ID] = token
}

var _ PasswordResetRepository = (*MockPasswordResetRepository)(nil)
