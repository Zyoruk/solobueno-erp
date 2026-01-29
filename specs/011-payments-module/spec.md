# Feature Specification: Payments Module

**Feature Branch**: `011-payments-module`  
**Created**: 2025-01-29  
**Status**: Draft  
**Dependencies**: 009-orders-module

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Customer Pays by Card (Priority: P1)

As a customer finishing my meal, I want to pay with my credit or debit card, so that I can complete my transaction conveniently.

**Why this priority**: Card payments are the most common payment method in modern restaurants.

**Independent Test**: Can be fully tested by processing a card payment through the payment terminal.

**Acceptance Scenarios**:

1. **Given** a customer is ready to pay, **When** the cashier selects card payment, **Then** the payment terminal activates for card input.

2. **Given** a card is presented, **When** the transaction processes, **Then** authorization is received within 10 seconds.

3. **Given** payment is successful, **When** confirmed, **Then** the order is marked as paid and a receipt option is offered.

4. **Given** payment is declined, **When** the response is received, **Then** the cashier sees a clear error message and can retry or select another method.

---

### User Story 2 - Customer Pays by Cash (Priority: P1)

As a customer, I want to pay with cash, so that I can use my preferred payment method.

**Why this priority**: Cash payments remain common and must be supported.

**Independent Test**: Can be fully tested by entering cash amount and verifying change calculation.

**Acceptance Scenarios**:

1. **Given** a customer pays with cash, **When** the cashier enters the amount tendered, **Then** the system calculates and displays change due.

2. **Given** the cash drawer is configured, **When** payment is completed, **Then** the drawer opens automatically.

3. **Given** exact change is given, **When** the transaction completes, **Then** the order is marked as paid.

---

### User Story 3 - Customer Splits Payment (Priority: P2)

As a customer dining with friends, I want to split the bill and pay with different methods, so that everyone can pay their share.

**Why this priority**: Split payments are frequently requested by groups.

**Independent Test**: Can be fully tested by paying for one order with multiple payment methods.

**Acceptance Scenarios**:

1. **Given** a split payment is initiated, **When** the cashier enters the first amount, **Then** the system shows remaining balance.

2. **Given** partial card payment is made, **When** processing, **Then** the system accepts and records the partial amount.

3. **Given** the remaining balance is paid in cash, **When** completed, **Then** both payments are recorded and the order is marked as paid.

---

### User Story 4 - Cashier Processes Refund (Priority: P2)

As a cashier, I want to process refunds for cancelled items or returned orders, so that customers receive their money back.

**Why this priority**: Refunds are necessary for customer service but less frequent than payments.

**Independent Test**: Can be fully tested by initiating a refund and verifying money is returned.

**Acceptance Scenarios**:

1. **Given** a manager authorizes a refund, **When** the cashier processes it, **Then** the refund is sent to the original payment method.

2. **Given** original payment was by card, **When** refund is processed, **Then** the card is credited and customer sees pending refund.

3. **Given** original payment was cash, **When** refund is processed, **Then** the cash drawer opens and amount is recorded.

4. **Given** a partial refund is needed, **When** processed, **Then** only the specified amount is refunded.

---

### User Story 5 - Manager Views Payment Reports (Priority: P3)

As a restaurant manager, I want to view payment summaries and reports, so that I can reconcile the cash drawer and track revenue.

**Why this priority**: Payment reporting is important for operations but not blocking for transactions.

**Independent Test**: Can be fully tested by viewing reports and verifying they match transaction records.

**Acceptance Scenarios**:

1. **Given** a manager opens payment reports, **When** selecting today's date, **Then** they see totals by payment method.

2. **Given** end-of-day reconciliation, **When** the manager views the report, **Then** they can compare expected vs actual cash.

3. **Given** a discrepancy exists, **When** investigating, **Then** the manager can view individual transactions.

---

### Edge Cases

- What happens when payment terminal is offline? System allows recording payment as "pending verification" with manager approval.
- What happens when a refund exceeds original payment? System prevents refund greater than charged amount.
- What happens during network timeout? Transaction status is verified before assuming failure.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST support card payments via payment terminal integration.

- **FR-002**: System MUST support cash payments with change calculation.

- **FR-003**: System MUST support split payments across multiple methods.

- **FR-004**: System MUST support partial payments with remaining balance tracking.

- **FR-005**: System MUST support full and partial refunds.

- **FR-006**: System MUST require manager authorization for refunds.

- **FR-007**: System MUST record all payment transactions with complete audit trail.

- **FR-008**: System MUST calculate tips and add to payment total when applicable.

- **FR-009**: System MUST support payment plugins for different processors (Stripe, local processors).

- **FR-010**: System MUST track payment status: pending, authorized, captured, refunded, failed.

- **FR-011**: System MUST publish payment events for notifications and analytics.

- **FR-012**: System MUST support opening cash drawer on cash transactions.

- **FR-013**: System MUST NOT store raw card data (PCI compliance).

### Key Entities

- **Payment**: Transaction recording money received; has order reference, amount, method, status, timestamp.
- **PaymentMethod**: Type of payment accepted; configured per tenant (card, cash, voucher, etc.).
- **Refund**: Return of payment; has original payment reference, amount, reason, authorizing user.
- **PaymentTerminal**: Physical device for card processing; has status, assigned location.
- **CashDrawer**: Physical drawer for cash; tracked for reconciliation.
- **PaymentPlugin**: Integration with payment processor (Stripe, etc.).

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Card payments process and return response within 10 seconds.

- **SC-002**: Cash change calculation is accurate to the cent 100% of the time.

- **SC-003**: Split payments correctly allocate amounts across methods.

- **SC-004**: 100% of payment transactions are recorded with complete audit trail.

- **SC-005**: Refunds process within 15 seconds with customer confirmation.

- **SC-006**: End-of-day reports match actual payment records with zero discrepancies.

- **SC-007**: System handles payment terminal disconnection gracefully with clear user feedback.
