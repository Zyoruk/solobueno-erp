import { describe, it, expect } from 'vitest';
import type {
  BaseEntity,
  TenantEntity,
  User,
  UserRole,
  OrderStatus,
  TableStatus,
  PaymentMethod,
} from './index';

describe('@solobueno/types', () => {
  describe('BaseEntity', () => {
    it('should define required fields', () => {
      const entity: BaseEntity = {
        id: '123',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };
      expect(entity.id).toBe('123');
      expect(entity.createdAt).toBeDefined();
      expect(entity.updatedAt).toBeDefined();
    });
  });

  describe('TenantEntity', () => {
    it('should extend BaseEntity with tenantId', () => {
      const entity: TenantEntity = {
        id: '123',
        tenantId: 'tenant-456',
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };
      expect(entity.tenantId).toBe('tenant-456');
    });
  });

  describe('User', () => {
    it('should define user with all required fields', () => {
      const user: User = {
        id: 'user-1',
        tenantId: 'tenant-1',
        email: 'test@example.com',
        firstName: 'John',
        lastName: 'Doe',
        role: 'admin',
        isActive: true,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      };
      expect(user.email).toBe('test@example.com');
      expect(user.role).toBe('admin');
    });

    it('should support all user roles', () => {
      const roles: UserRole[] = [
        'owner',
        'admin',
        'manager',
        'cashier',
        'waiter',
        'kitchen',
        'viewer',
      ];
      expect(roles).toHaveLength(7);
    });
  });

  describe('Order', () => {
    it('should support all order statuses', () => {
      const statuses: OrderStatus[] = [
        'draft',
        'sent',
        'in_progress',
        'ready',
        'served',
        'paid',
        'cancelled',
      ];
      expect(statuses).toHaveLength(7);
    });
  });

  describe('Table', () => {
    it('should support all table statuses', () => {
      const statuses: TableStatus[] = ['available', 'occupied', 'reserved', 'cleaning'];
      expect(statuses).toHaveLength(4);
    });
  });

  describe('Payment', () => {
    it('should support all payment methods', () => {
      const methods: PaymentMethod[] = ['cash', 'card', 'voucher'];
      expect(methods).toHaveLength(3);
    });
  });
});
