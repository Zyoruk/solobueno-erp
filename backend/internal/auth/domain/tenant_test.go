package domain

import "testing"

func TestTenant_IsOperational(t *testing.T) {
	t1 := Tenant{IsActive: true}
	if !t1.IsOperational() {
		t.Error("Active tenant should be operational")
	}

	t2 := Tenant{IsActive: false}
	if t2.IsOperational() {
		t.Error("Inactive tenant should not be operational")
	}
}

func TestTenant_TableName(t *testing.T) {
	tenant := Tenant{}
	if tenant.TableName() != "tenants" {
		t.Errorf("TableName() = %q, want %q", tenant.TableName(), "tenants")
	}
}
