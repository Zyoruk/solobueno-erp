# Feature Specification: Tables Module

**Feature Branch**: `008-tables-module`  
**Created**: 2025-01-29  
**Status**: Draft  
**Dependencies**: 004-auth-module, 005-config-module

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Waiter Selects a Table to Start Order (Priority: P1)

As a waiter, I want to select a table from the floor plan to start taking an order, so that the order is associated with the correct table.

**Why this priority**: Table selection is the entry point for dine-in order taking.

**Independent Test**: Can be fully tested by selecting a table and starting an order.

**Acceptance Scenarios**:

1. **Given** a waiter opens the table view, **When** the screen loads, **Then** they see a visual representation of all tables with their current status.

2. **Given** a table is available, **When** the waiter taps on it, **Then** a new session starts and they can begin adding items.

3. **Given** a table has an active session, **When** the waiter taps on it, **Then** they see the current order and can add more items.

4. **Given** tables have different statuses, **When** displayed, **Then** they are color-coded (green=available, yellow=occupied, red=needs attention).

---

### User Story 2 - Waiter Views Table Status at a Glance (Priority: P1)

As a waiter managing multiple tables, I want to see the status of all my tables at a glance, so that I can prioritize my attention.

**Why this priority**: Quick status overview is essential for efficient floor management.

**Independent Test**: Can be fully tested by checking table statuses match their actual states.

**Acceptance Scenarios**:

1. **Given** a table has been waiting for food for 20+ minutes, **When** viewing the floor plan, **Then** the table shows a visual alert indicator.

2. **Given** a table is ready for the check, **When** the customer requests it, **Then** the waiter marks it and the status updates immediately.

3. **Given** multiple tables need attention, **When** viewing the floor plan, **Then** tables are sorted/highlighted by urgency.

---

### User Story 3 - Host Manages Reservations (Priority: P2)

As a host, I want to manage table reservations, so that customers can book ahead and we can prepare for their arrival.

**Why this priority**: Reservations are important for planning but restaurants can operate without them.

**Independent Test**: Can be fully tested by creating a reservation and seating the party.

**Acceptance Scenarios**:

1. **Given** a customer calls to reserve, **When** the host creates a reservation with name, time, and party size, **Then** the reservation appears in the system.

2. **Given** a reservation time approaches, **When** viewing the floor plan, **Then** the reserved table shows the upcoming reservation details.

3. **Given** a reserved party arrives, **When** the host seats them, **Then** the reservation is marked as seated and the table session begins.

4. **Given** a reservation is cancelled, **When** the host removes it, **Then** the table becomes available for walk-ins.

---

### User Story 4 - Manager Configures Floor Layout (Priority: P2)

As a restaurant manager, I want to configure the floor layout with table positions and capacities, so that the digital floor plan matches the physical restaurant.

**Why this priority**: Layout configuration is required setup but changes infrequently.

**Independent Test**: Can be fully tested by arranging tables and verifying the layout saves correctly.

**Acceptance Scenarios**:

1. **Given** a manager opens layout editor, **When** they drag and position tables, **Then** the positions are saved and reflected for all staff.

2. **Given** a manager adds a new table, **When** they specify the table number and capacity, **Then** the table appears in the floor plan.

3. **Given** a table is removed from the floor, **When** it has no active session, **Then** it can be deleted or marked inactive.

4. **Given** a table needs to combine with another, **When** the manager links them, **Then** they show as a combined table with summed capacity.

---

### User Story 5 - Waiter Transfers or Merges Tables (Priority: P3)

As a waiter, I want to transfer an order to another table or merge tables, so that I can accommodate customer requests.

**Why this priority**: Table operations are helpful but not required for basic functionality.

**Independent Test**: Can be fully tested by transferring an order and verifying it moves correctly.

**Acceptance Scenarios**:

1. **Given** a customer wants to move, **When** the waiter transfers the order to another table, **Then** all items move to the new table and the old table becomes available.

2. **Given** two parties want to combine, **When** the waiter merges the tables, **Then** orders are combined into one check.

3. **Given** a merged table wants to split, **When** the waiter separates them, **Then** the system allows items to be divided between tables.

---

### Edge Cases

- What happens when a table is selected while being cleaned? Show "Table being cleaned" status with estimated ready time.
- What happens when reservations overlap? Prevent booking and suggest alternative times.
- What happens when a server's shift ends? Tables can be reassigned to another server.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST support defining tables with number, capacity, position, and shape.

- **FR-002**: System MUST display tables in a visual floor plan layout.

- **FR-003**: System MUST track table status: available, occupied, reserved, cleaning, combined.

- **FR-004**: System MUST support table sessions that track the duration and current order.

- **FR-005**: System MUST support reservations with customer name, time, party size, and notes.

- **FR-006**: System MUST show visual indicators for table status with color coding.

- **FR-007**: System MUST support combining/merging tables for large parties.

- **FR-008**: System MUST support transferring orders between tables.

- **FR-009**: System MUST support assigning tables to specific servers.

- **FR-010**: System MUST track time elapsed since table was seated.

- **FR-011**: System MUST alert when tables exceed expected dining time.

- **FR-012**: Table status changes MUST sync across devices within 5 seconds.

- **FR-013**: System MUST support multiple floors/sections for larger restaurants.

### Key Entities

- **Table**: Physical table in the restaurant; has number, capacity, position coordinates, shape, section.
- **TableSession**: Active period from customer seating to checkout; links to orders.
- **Reservation**: Future booking for a table; has customer info, date/time, party size, status.
- **Section**: Area of the restaurant (patio, main floor, bar); contains tables.
- **ServerAssignment**: Mapping of server (user) to tables for their shift.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Floor plan loads within 1 second showing all tables with current status.

- **SC-002**: Table status changes sync across all devices within 5 seconds.

- **SC-003**: Waiters can start a new order from table selection within 3 taps.

- **SC-004**: Reservation conflicts are detected and prevented 100% of the time.

- **SC-005**: Table layouts support up to 100 tables per restaurant without performance issues.

- **SC-006**: Time tracking accuracy within 1 minute for session duration.

- **SC-007**: Floor plan works on tablets in both portrait and landscape orientation.
