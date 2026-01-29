/**
 * @solobueno/types - Shared TypeScript Types
 *
 * Domain entity types shared across all applications.
 * Ensures type consistency between frontend and backend.
 *
 * @packageDocumentation
 */

// Base types
export interface BaseEntity {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface TenantEntity extends BaseEntity {
  tenantId: string;
}

// User & Auth types
export type UserRole = 'owner' | 'admin' | 'manager' | 'cashier' | 'waiter' | 'kitchen' | 'viewer';

export interface User extends TenantEntity {
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
  isActive: boolean;
}

// Menu types
export interface MenuItem extends TenantEntity {
  name: string;
  description: string;
  price: number;
  categoryId: string;
  imageUrl?: string;
  isAvailable: boolean;
}

export interface MenuCategory extends TenantEntity {
  name: string;
  description?: string;
  displayOrder: number;
  iconName?: string;
}

// Order types
export type OrderStatus = 'draft' | 'sent' | 'in_progress' | 'ready' | 'served' | 'paid' | 'cancelled';

export interface Order extends TenantEntity {
  tableId?: string;
  userId: string;
  status: OrderStatus;
  subtotal: number;
  tax: number;
  total: number;
  notes?: string;
}

export interface OrderItem extends BaseEntity {
  orderId: string;
  menuItemId: string;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
  notes?: string;
}

// Table types
export type TableStatus = 'available' | 'occupied' | 'reserved' | 'cleaning';

export interface Table extends TenantEntity {
  number: string;
  capacity: number;
  status: TableStatus;
  sectionId?: string;
  positionX: number;
  positionY: number;
}

// Payment types
export type PaymentMethod = 'cash' | 'card' | 'voucher';
export type PaymentStatus = 'pending' | 'authorized' | 'captured' | 'refunded' | 'failed';

export interface Payment extends TenantEntity {
  orderId: string;
  amount: number;
  method: PaymentMethod;
  status: PaymentStatus;
  reference?: string;
}
