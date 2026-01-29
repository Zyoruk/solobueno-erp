# Feature Specification: Billing Module

**Feature Branch**: `012-billing-module`  
**Created**: 2025-01-29  
**Status**: Draft  
**Dependencies**: 009-orders-module, 011-payments-module

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Customer Receives Invoice (Priority: P1)

As a customer completing payment, I want to receive a proper invoice, so that I have documentation of my purchase for personal or business records.

**Why this priority**: Invoice generation is legally required in Costa Rica and most jurisdictions.

**Independent Test**: Can be fully tested by completing a payment and verifying invoice is generated.

**Acceptance Scenarios**:

1. **Given** payment is completed, **When** the transaction finalizes, **Then** an invoice is automatically generated with all required legal information.

2. **Given** an invoice is generated, **When** the customer requests it, **Then** they can receive it via email or printed receipt.

3. **Given** the invoice has customer details, **When** they provide tax ID, **Then** the invoice includes the tax identification for deduction purposes.

---

### User Story 2 - System Generates Compliant Invoices (Priority: P1)

As a restaurant owner in Costa Rica, I want invoices to comply with Hacienda (tax authority) requirements, so that my business meets fiscal obligations.

**Why this priority**: Non-compliant invoices can result in fines and legal issues.

**Independent Test**: Can be fully tested by generating invoices and verifying they meet Hacienda format requirements.

**Acceptance Scenarios**:

1. **Given** the Costa Rica billing plugin is active, **When** an invoice is generated, **Then** it includes all fields required by Ministerio de Hacienda.

2. **Given** an invoice is created, **When** submitted to Hacienda, **Then** the system receives and stores the authorization code.

3. **Given** Hacienda rejects an invoice, **When** the rejection is received, **Then** the system alerts the manager with the specific error.

4. **Given** the tax authority is unreachable, **When** generating invoices, **Then** they are queued and retried automatically.

---

### User Story 3 - Manager Views Invoice History (Priority: P2)

As a manager or accountant, I want to view and search invoice history, so that I can reconcile records and respond to customer inquiries.

**Why this priority**: Invoice lookup is needed for operations but less frequent than generation.

**Independent Test**: Can be fully tested by searching for invoices by various criteria.

**Acceptance Scenarios**:

1. **Given** a manager searches by invoice number, **When** found, **Then** they see the complete invoice with all details.

2. **Given** a date range is specified, **When** searching, **Then** all invoices in that range are displayed.

3. **Given** a customer calls about an invoice, **When** the manager finds it, **Then** they can resend or reprint it.

---

### User Story 4 - System Calculates Taxes Correctly (Priority: P1)

As a restaurant owner, I want taxes calculated correctly on every order, so that I collect the right amount and report accurately.

**Why this priority**: Tax calculation accuracy is legally required and financially critical.

**Independent Test**: Can be fully tested by creating orders with various items and verifying tax calculations.

**Acceptance Scenarios**:

1. **Given** items have different tax categories, **When** the order is totaled, **Then** each tax is calculated separately and shown as line items.

2. **Given** Costa Rican IVA (13%) applies, **When** calculated, **Then** the tax amount is exactly 13% of the taxable subtotal.

3. **Given** service charge (10%) is configured, **When** applied, **Then** it is calculated and shown separately from taxes.

4. **Given** some items are tax-exempt, **When** on the same order, **Then** taxes are calculated only on taxable items.

---

### User Story 5 - Manager Voids Invoice (Priority: P2)

As a manager, I want to void an invoice when an order is cancelled or incorrectly processed, so that records are accurate.

**Why this priority**: Voiding is necessary for corrections but requires authorization.

**Independent Test**: Can be fully tested by voiding an invoice and verifying the void is recorded.

**Acceptance Scenarios**:

1. **Given** an invoice needs to be voided, **When** the manager authorizes and processes the void, **Then** a credit note or void record is created.

2. **Given** the invoice was submitted to Hacienda, **When** voided, **Then** the void is also reported to the tax authority.

3. **Given** an invoice is voided, **When** viewing reports, **Then** the voided invoice is excluded from revenue totals but visible in audit.

---

### Edge Cases

- What happens when Hacienda API is down for extended period? Invoices queue locally with offline mode warning.
- What happens when tax rates change? New rate applies to new orders; historical invoices retain original rate.
- What happens when generating invoice for international customer? System supports invoices without local tax ID.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST generate invoices automatically upon payment completion.

- **FR-002**: System MUST support billing plugins for different jurisdictions (Costa Rica, generic).

- **FR-003**: Costa Rica plugin MUST comply with Ministerio de Hacienda electronic invoicing (FE) requirements.

- **FR-004**: System MUST calculate IVA (13%) and service charge (10%) for Costa Rica.

- **FR-005**: System MUST support configurable tax rates per tenant/jurisdiction.

- **FR-006**: System MUST support different tax categories per menu item.

- **FR-007**: Invoices MUST include: business info, customer info (optional), items, taxes, totals, unique number.

- **FR-008**: System MUST store all invoices with complete audit trail.

- **FR-009**: System MUST support invoice void/cancellation with authorization.

- **FR-010**: System MUST support credit notes for partial refunds.

- **FR-011**: System MUST support sending invoices via email.

- **FR-012**: System MUST support printing invoices to receipt printer.

- **FR-013**: System MUST queue invoices for fiscal authority when offline and retry automatically.

- **FR-014**: System MUST publish invoice events for notifications and analytics.

### Key Entities

- **Invoice**: Official document for a transaction; has number, items, taxes, totals, customer info, fiscal status.
- **InvoiceLine**: Single item on an invoice; has description, quantity, unit price, tax category, total.
- **TaxRate**: Configuration for a tax type; has name, percentage, applicable categories.
- **FiscalSubmission**: Record of invoice submission to tax authority; has status, response, authorization code.
- **CreditNote**: Document reversing part or all of an invoice; linked to original invoice.
- **BillingPlugin**: Integration with fiscal authority (Costa Rica Hacienda, generic, etc.).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Invoices generate within 5 seconds of payment completion.

- **SC-002**: 100% of invoices meet fiscal authority format requirements.

- **SC-003**: Tax calculations are accurate to the cent on 100% of invoices.

- **SC-004**: Fiscal authority submissions succeed on first attempt 99% of the time.

- **SC-005**: Failed submissions retry automatically and succeed within 24 hours.

- **SC-006**: Invoice search returns results within 2 seconds for queries spanning 1 year.

- **SC-007**: Voided invoices are properly excluded from revenue reports while remaining in audit trail.
