<!--
SYNC IMPACT REPORT
==================
Version change: 1.6.1 → 1.7.0 (MINOR - all technology decisions finalized)

Modified sections:
- Stack Requirements → locked in specific technologies
- i18n Strategy → package-driven selected
- Infrastructure → AWS Lightsail + us-east-1 confirmed

Finalized decisions:
- Cloud: AWS Lightsail (us-east-1)
- i18n: Package-driven (@solobueno/i18n)
- CI/CD: GitHub Actions
- Email: AWS SES
- Secrets: AWS SSM Parameter Store
- Migrations: golang-migrate
- REST: Chi
- Logger: zerolog
- Job Queue: Asynq
- Monorepo: Turborepo
- Registry: Amazon ECR
- Event Bus (prod): NATS
- Offline Sync: WatermelonDB
- Search: PostgreSQL FTS (Phase 1)

Templates requiring updates:
- .specify/templates/plan-template.md ✅ (no changes needed - generic template)
- .specify/templates/spec-template.md ✅ (no changes needed - generic template)
- .specify/templates/tasks-template.md ✅ (no changes needed - generic template)
- .specify/templates/checklist-template.md ✅ (no changes needed - generic template)
- .specify/templates/agent-file-template.md ✅ (no changes needed - generic template)

Follow-up TODOs: None
==================
-->

# Solobueno ERP Constitution

## Core Principles

### I. Mobile-First for Operations

Order taking and table management MUST be designed for mobile devices as the primary interface.

- Mobile app MUST be the reference implementation for waiter/floor staff workflows
- Touch interactions MUST be optimized for one-handed smartphone operation
- Screen layouts MUST prioritize portrait orientation for order-taking flows
- Desktop/tablet interfaces MAY adapt mobile patterns but MUST NOT drive mobile design
- Critical paths (new order, add item, send to kitchen, close table) MUST work on 5" screens

**Rationale**: Restaurant floor staff move constantly. Smartphones are more practical than tablets
for mobility, and designing mobile-first ensures usability under real service conditions.

### II. Domain-Driven & Event-Driven Design

The system MUST be architected as a modular monolith with clear bounded contexts and event-driven
communication, ready for microservice extraction when scale demands.

#### Bounded Contexts (as internal modules)

| Module          | Domain Responsibility                                            |
| --------------- | ---------------------------------------------------------------- |
| `auth`          | Authentication, authorization, tenant context, user management   |
| `menu`          | Menu items, categories, modifiers, pricing rules                 |
| `orders`        | Order lifecycle, cart, kitchen tickets, order history            |
| `tables`        | Table layout, reservations, active sessions                      |
| `inventory`     | Stock levels, ingredients, low-stock alerts                      |
| `billing`       | Invoice generation, fiscal plugin orchestration, tax calculation |
| `payments`      | Payment processing, plugin orchestration, refunds                |
| `reporting`     | Analytics, reports, data exports                                 |
| `config`        | Tenant configuration, feature flags, branding                    |
| `feedback`      | Customer complaints, ratings, staff performance tracking         |
| `analytics`     | Clickstream collection, business intelligence events             |
| `notifications` | Push notifications, email, SMS, in-app alerts                    |
| `audit`         | Action logging, compliance trail, change history                 |
| `jobs`          | Background tasks, scheduled jobs, async processing               |
| `media`         | File uploads, image processing, CDN management                   |
| `search`        | Full-text search, indexing, autocomplete (optional module)       |

#### Module Communication Rules

**Synchronous (Interface Calls)** - Use when:

- Caller MUST have an immediate response to proceed
- Query operations (read data from another module)
- Validation that blocks the current operation

**Asynchronous (Domain Events)** - Use when:

- Side effects that don't affect the caller's response
- Multiple modules need to react to the same action
- Operations that can tolerate eventual consistency
- Decoupling is more important than immediate consistency

| Communication Type          | Pattern         | Example                                    |
| --------------------------- | --------------- | ------------------------------------------ |
| Query another module        | Interface call  | Orders calls `menu.GetItem()` to get price |
| Validate before action      | Interface call  | Orders calls `inventory.CheckStock()`      |
| Notify something happened   | Domain event    | Orders publishes `OrderCompleted`          |
| Trigger side effects        | Domain event    | Inventory subscribes to `OrderCompleted`   |
| Complex multi-step workflow | Saga (events)   | Payment → Invoice → Notification           |
| Track user behavior         | Analytics event | UI publishes `MenuItemViewed`              |

#### Domain Events as First-Class Citizens

- Every significant state change MUST publish a domain event
- Events MUST be named in past tense (something that happened): `OrderCreated`, `PaymentReceived`
- Events MUST be immutable and contain all data needed by consumers
- Events MUST include: `EventID`, `TenantID`, `Timestamp`, `AggregateID`, `Payload`
- Consumers MUST be idempotent (handle duplicate events gracefully)

**Core Domain Events** (initial, will grow):

| Module          | Events Published                                                                           |
| --------------- | ------------------------------------------------------------------------------------------ |
| `orders`        | `OrderCreated`, `OrderItemAdded`, `OrderSentToKitchen`, `OrderCompleted`, `OrderCancelled` |
| `tables`        | `TableSessionStarted`, `TableSessionClosed`, `ReservationCreated`, `ReservationCancelled`  |
| `payments`      | `PaymentInitiated`, `PaymentSucceeded`, `PaymentFailed`, `RefundProcessed`                 |
| `billing`       | `InvoiceGenerated`, `InvoiceSent`, `InvoiceVoided`                                         |
| `inventory`     | `StockDepleted`, `LowStockAlert`, `StockReplenished`                                       |
| `menu`          | `MenuItemCreated`, `MenuItemUpdated`, `MenuItemDisabled`, `PriceChanged`                   |
| `auth`          | `UserCreated`, `UserRoleChanged`, `TenantCreated`, `LoginSucceeded`, `LoginFailed`         |
| `feedback`      | `ComplaintFiled`, `ComplaintResolved`, `RatingSubmitted`, `StaffRated`                     |
| `notifications` | `NotificationSent`, `NotificationFailed`, `NotificationRead`                               |
| `audit`         | `AuditEventRecorded` (internal, not for subscription)                                      |
| `jobs`          | `JobScheduled`, `JobStarted`, `JobCompleted`, `JobFailed`                                  |
| `media`         | `FileUploaded`, `FileDeleted`, `ImageProcessed`                                            |

**Analytics/Clickstream Events** (for business intelligence):

| Event                 | Purpose                       | Key Data                                                  |
| --------------------- | ----------------------------- | --------------------------------------------------------- |
| `MenuItemViewed`      | Track product interest        | `itemId`, `categoryId`, `userId`, `sessionId`, `duration` |
| `MenuCategoryBrowsed` | Track category popularity     | `categoryId`, `userId`, `sessionId`, `itemsViewed`        |
| `ItemAddedToCart`     | Track purchase intent         | `itemId`, `quantity`, `orderId`, `userId`                 |
| `ItemRemovedFromCart` | Track abandonment             | `itemId`, `quantity`, `orderId`, `reason`                 |
| `SearchPerformed`     | Track search behavior         | `query`, `resultsCount`, `userId`, `sessionId`            |
| `ModifierSelected`    | Track customization patterns  | `itemId`, `modifierId`, `userId`                          |
| `OrderAbandoned`      | Track incomplete orders       | `orderId`, `itemsInCart`, `totalValue`, `stage`           |
| `PromotionViewed`     | Track marketing effectiveness | `promotionId`, `userId`, `sessionId`                      |
| `PromotionApplied`    | Track promotion conversion    | `promotionId`, `orderId`, `discountAmount`                |
| `PageViewed`          | General navigation tracking   | `pageName`, `userId`, `sessionId`, `referrer`             |
| `SessionStarted`      | Track user sessions           | `sessionId`, `userId`, `deviceType`, `appVersion`         |
| `SessionEnded`        | Session duration tracking     | `sessionId`, `duration`, `pagesViewed`, `ordersPlaced`    |

#### Event Bus Abstraction

The event bus MUST be abstracted to allow swapping implementations:

```
┌─────────────────────────────────────────────────────────────┐
│                    EventBus Interface                        │
│  Publish(ctx, event) / Subscribe(eventType, handler)        │
└─────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  In-Memory   │ │    NATS      │ │    Kafka     │
    │  (Dev/Start) │ │  (Scale-up)  │ │  (Extract)   │
    └──────────────┘ └──────────────┘ └──────────────┘
```

- **Phase 1 (Monolith)**: In-memory event bus (synchronous or goroutine-based)
- **Phase 2 (Scale)**: NATS or Redis Streams for persistence and replay
- **Phase 3 (Microservices)**: Kafka/RabbitMQ when services are extracted

#### Event Design Rules

- Events MUST NOT contain references to domain objects (only IDs and value copies)
- Events MUST be backward-compatible (additive changes only, no field removal)
- Events MUST have a schema version for evolution
- Large payloads SHOULD use "thin event + fetch" pattern (event contains ID, consumer fetches details)
- Sensitive data (PII, payment details) MUST NOT appear in events
- Analytics events SHOULD be fire-and-forget (non-blocking)

#### Saga Pattern for Complex Workflows

Multi-step processes spanning modules MUST use the Saga pattern:

```
Example: Complete Order Saga

1. OrderService publishes OrderCompleted
2. PaymentService subscribes → processes payment → publishes PaymentSucceeded/Failed
3. BillingService subscribes to PaymentSucceeded → generates invoice → publishes InvoiceGenerated
4. InventoryService subscribes to OrderCompleted → deducts stock
5. NotificationService subscribes to InvoiceGenerated → sends receipt

Compensation (if payment fails):
- PaymentFailed triggers → OrderService marks order as payment_failed
- No invoice generated, no stock deducted
```

- Sagas MUST define compensation actions for failure scenarios
- Saga state MUST be persisted for recovery after crashes
- Long-running sagas MUST have timeout and retry policies

#### Microservice Extraction Criteria

A module SHOULD be extracted to a microservice only when:

- It has different scaling requirements than the monolith
- It needs independent deployment for release velocity
- Team ownership boundaries require process isolation
- Performance isolation is required (e.g., reporting heavy queries)

**Extraction is straightforward because**:

- Module interfaces become gRPC/HTTP clients
- In-memory event bus becomes message broker topic
- No business logic changes required

**Rationale**: Event-driven architecture decouples modules at design time, making runtime
decoupling (microservices) a deployment decision rather than a rewrite. Starting with a
modular monolith avoids distributed system complexity while keeping all options open.

### III. API-First Architecture

All functionality MUST be exposed through well-defined APIs before any UI implementation.

- **GraphQL**: Primary API for all client applications
  - Single GraphQL server with domain-organized resolvers
  - Schema MUST be the contract; generated types for clients
  - Subscriptions for real-time features (order status, kitchen display)
  - Subscriptions SHOULD be backed by domain events internally
- **REST**: Available for third-party integrations and webhooks
  - OpenAPI 3.0 specification required for REST endpoints
  - Webhooks MUST be triggered by domain events
- Internal modules communicate via Go interfaces and domain events, NOT HTTP/RPC
- API layer MUST be thin—delegate to domain modules, not contain business logic

**Rationale**: GraphQL provides flexibility for diverse clients. Domain events power real-time
subscriptions and webhooks uniformly. Clean boundaries enable future extraction.

### IV. Offline-First Operations

The system MUST function without internet connectivity for core restaurant operations.

- Order taking, kitchen display, and table management MUST work offline
- Local data storage MUST sync automatically when connectivity resumes
- Conflict resolution strategy MUST be defined (last-write-wins with server authority)
- Critical operations (orders, payments) MUST queue and retry on reconnection
- Mobile app MUST pre-cache menu, tables, and configuration on startup
- Offline actions SHOULD generate local events that sync to server event log
- Analytics events SHOULD queue locally and batch-upload when online

**Rationale**: Restaurants experience unreliable connectivity. A non-functional system during
service hours is unacceptable for any business in any country.

### V. Plugin-Driven Architecture

Country-specific, integration-specific, and business-specific features MUST be implemented as plugins.

- **Billing/Invoicing**: MUST be a plugin interface (Costa Rica Hacienda FE, other fiscal systems)
- **Payment Processing**: MUST be a plugin interface (Stripe, local processors)
- **Tax Calculation**: MUST be a plugin interface (VAT rates, service charges by jurisdiction)
- **Reporting**: MUST support plugin-based report generators for regulatory compliance
- **Analytics Export**: MAY be a plugin for sending events to external systems (Mixpanel, Amplitude)
- Core modules MUST define stable plugin contracts (Go interfaces)
- Plugins MUST be registrable per tenant without redeployment (configuration-driven)
- Plugins MAY subscribe to domain events for reactive integrations
- Plugin implementations MAY live in the monorepo or as external packages

**Rationale**: Expanding to multiple countries and businesses requires swappable compliance,
payment, and reporting modules. Hardcoding country-specific logic prevents growth.

### VI. White-Label & Multi-Tenant

The system MUST support multiple independent businesses from a single deployment.

- Each tenant MUST have isolated data (no cross-tenant data leakage)
- Tenant context MUST be established at request entry and propagated through all modules
- Domain events MUST include TenantID for proper routing and isolation
- Branding (logo, colors, name) MUST be configurable per tenant
- Feature flags MUST control functionality per tenant (enable/disable modules)
- Menu structure, pricing, tax rules, and workflows MUST be tenant-configurable
- Plugin selection MUST be configurable per tenant
- Tenant configuration MUST NOT require code changes or redeployment

**Rationale**: SaaS model enables serving multiple restaurants efficiently. White-labeling
allows reselling to chains or franchises with their own branding.

### VII. Type Safety

All code MUST use static typing with strict compiler/linter enforcement.

- **Go**: Leverage Go's native type system; avoid `interface{}` without justification
- **TypeScript**: Strict mode required; no `any` types without documented justification
- GraphQL schema MUST generate typed clients (codegen for React Native and web)
- Database schemas MUST have corresponding type definitions (sqlc or similar for Go)
- Domain types (`Money`, `OrderID`, etc.) MUST be distinct types, not primitives
- Domain events MUST have strongly-typed definitions (no generic maps)

**Rationale**: ERP systems handle money, inventory, and compliance data. Runtime type errors
in production cause financial discrepancies and break trust.

### VIII. Test-Driven Development

Business-critical paths MUST have tests written before implementation.

- Payment processing, order lifecycle, and inventory mutations require tests first
- Plugin contracts MUST have compliance test suites plugins must pass
- Module interfaces MUST have contract tests verifying behavior
- Event handlers MUST have tests verifying idempotency
- Saga workflows MUST have tests covering success and compensation paths
- Integration tests MUST cover cross-module event-driven workflows
- E2E tests MUST cover critical user journeys (create order → kitchen → payment)
- Test coverage for business logic MUST exceed 80%

**Rationale**: Multi-tenant ERP systems are business-critical for every customer. Bugs affect
multiple businesses simultaneously; rigorous testing is non-negotiable.

### IX. Internationalization & Pluggable Compliance

The system MUST support multiple languages and regional compliance through configuration and plugins.

**i18n Strategy: Package-Driven** (decided)

The `@solobueno/i18n` package in the monorepo manages all translations:

```
packages/i18n/
├── src/
│   ├── locales/
│   │   ├── es-419.json    # Spanish (Latin America) - Primary
│   │   └── en.json        # English
│   ├── index.ts           # Export functions
│   └── types.ts           # Type-safe keys
├── package.json
└── README.md
```

**Why Package-Driven**:

- Simpler architecture for MVP
- Type-safe translation keys (compile-time checking)
- Faster load times (bundled, no API calls)
- Can migrate to database-driven later if tenant-specific overrides needed

**Requirements**:

- All user-facing strings MUST use i18n keys (no hardcoded text in UI)
- Default languages: Spanish (Latin America), English
- Additional languages MUST be addable without code changes
- Currency MUST be configurable per tenant (CRC, USD, EUR, etc.)
- Date/time MUST respect tenant timezone configuration
- Number formatting MUST follow tenant locale

**Compliance**: Tax/fiscal compliance MUST be handled by jurisdiction plugins (not hardcoded).

**Rationale**: International expansion requires flexible localization. Database-driven allows
business users to update translations; package-driven is simpler for developer-managed i18n.

### X. User-Centric Interface Design

All interfaces MUST prioritize speed and simplicity for restaurant staff under pressure.

- Mobile: Critical actions MUST be reachable in ≤2 taps from any screen
- Mobile: Large touch targets (minimum 44pt) for error-free tapping while moving
- Backoffice MUST load actionable dashboards in <3 seconds
- Real-time updates (via GraphQL subscriptions) MUST reflect order/kitchen status within 1 second
- Error messages MUST be localized and actionable (not technical codes)
- Color coding and icons MUST convey status without requiring text reading
- Accessibility: MUST support dynamic text sizing and high contrast modes

**Rationale**: Restaurant staff work in high-pressure, fast-paced environments across all
countries and cultures. Complex UIs cause errors, slow service, and staff frustration.

### XI. Observability & Monitoring

The system MUST have comprehensive observability through structured logging, with adapter patterns
enabling future expansion to metrics and tracing.

#### Logging

**Logger Adapter Pattern**:

```go
// Logger interface - implementations can be swapped
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    With(fields ...Field) Logger  // Create child logger with context
}

// Field represents a structured log field
type Field struct {
    Key   string
    Value any
}
```

**Logging Requirements**:

- All logs MUST be structured (JSON in production, human-readable in dev)
- All logs MUST include: `timestamp`, `level`, `message`, `tenant_id`, `request_id`
- Sensitive data (passwords, tokens, PII) MUST NOT appear in logs
- Log levels MUST be configurable per environment
- Logger implementation MUST be injectable (adapter pattern)

**Standard Log Fields**:

| Field         | Description                            | Required                    |
| ------------- | -------------------------------------- | --------------------------- |
| `timestamp`   | ISO 8601 format                        | Yes                         |
| `level`       | debug, info, warn, error               | Yes                         |
| `message`     | Human-readable description             | Yes                         |
| `tenant_id`   | Current tenant context                 | Yes (if in tenant context)  |
| `request_id`  | Correlation ID for request tracing     | Yes (if in request context) |
| `user_id`     | Acting user                            | Yes (if authenticated)      |
| `module`      | Source module (orders, payments, etc.) | Yes                         |
| `duration_ms` | Operation duration                     | For operations              |
| `error`       | Error details                          | For errors                  |

**Log Levels by Environment**:

| Environment | Default Level | Notes                         |
| ----------- | ------------- | ----------------------------- |
| dev         | debug         | All logs visible              |
| test        | warn          | Reduce noise in CI            |
| staging     | info          | Production-like               |
| prod        | info          | Consider warn for high-volume |

**Logger Implementations** (swap via configuration):

| Phase   | Implementation      | Use Case                    |
| ------- | ------------------- | --------------------------- |
| Phase 1 | `zerolog` / `zap`   | Local structured logging    |
| Phase 2 | + File/stdout       | Container log collection    |
| Phase 3 | + Loki / CloudWatch | Centralized log aggregation |

#### Future Observability (prepare interfaces, implement later)

**Metrics Adapter** (for future):

```go
type Metrics interface {
    Counter(name string, tags ...Tag) Counter
    Gauge(name string, tags ...Tag) Gauge
    Histogram(name string, tags ...Tag) Histogram
}
```

**Tracing Adapter** (for future):

```go
type Tracer interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
}
```

- Metrics and tracing interfaces SHOULD be defined but MAY have no-op implementations initially
- When needed, implement with Prometheus (metrics) and OpenTelemetry (tracing)

#### What to Log

| Category            | Log Level | Example                                                                         |
| ------------------- | --------- | ------------------------------------------------------------------------------- |
| Request received    | info      | `"message": "GraphQL request", "operation": "CreateOrder"`                      |
| Request completed   | info      | `"message": "Request completed", "duration_ms": 45, "status": "success"`        |
| Business event      | info      | `"message": "Order sent to kitchen", "order_id": "...", "items_count": 5`       |
| External API call   | info      | `"message": "Payment API called", "provider": "stripe", "duration_ms": 230`     |
| Validation failure  | warn      | `"message": "Invalid input", "field": "email", "reason": "malformed"`           |
| Recoverable error   | warn      | `"message": "Retry succeeded", "attempt": 2, "operation": "sync"`               |
| Unrecoverable error | error     | `"message": "Payment failed", "error": "insufficient_funds", "order_id": "..."` |
| Debug details       | debug     | `"message": "Cache hit", "key": "menu:tenant123"`                               |

#### Clickstream & Analytics Events

Analytics events track user behavior for business intelligence. They are separate from domain
events and optimized for analysis rather than system reactions.

**Analytics Event Flow**:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Mobile    │────►│  Analytics  │────►│   Event     │────►│  Reporting  │
│     UI      │     │   Module    │     │   Store     │     │   Module    │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼ (optional plugin)
                    ┌─────────────┐
                    │  External   │
                    │  (Mixpanel) │
                    └─────────────┘
```

**Analytics Collection Requirements**:

- Analytics events MUST be non-blocking (fire-and-forget)
- Analytics events MUST queue locally when offline
- Analytics events SHOULD batch-upload (e.g., every 30 seconds or 50 events)
- Analytics events MUST include session context for funnel analysis
- User consent MUST be respected (configurable per tenant/user)

**Business Intelligence Queries Enabled**:

| Question                          | Events Used                                                        |
| --------------------------------- | ------------------------------------------------------------------ |
| Most viewed items                 | `MenuItemViewed` aggregated by `itemId`                            |
| Most purchased items              | `OrderCompleted` → items breakdown                                 |
| Conversion rate (view → purchase) | `MenuItemViewed` vs `ItemAddedToCart` vs `OrderCompleted`          |
| Cart abandonment rate             | `ItemAddedToCart` vs `OrderAbandoned`                              |
| Average order value               | `OrderCompleted` → `totalValue`                                    |
| Peak hours                        | `OrderCreated` by hour                                             |
| Popular modifiers                 | `ModifierSelected` aggregated                                      |
| Search effectiveness              | `SearchPerformed` → `resultsCount` vs subsequent `ItemAddedToCart` |
| Session duration                  | `SessionStarted` → `SessionEnded`                                  |
| Promotion effectiveness           | `PromotionViewed` vs `PromotionApplied`                            |

#### Feedback Module

The `feedback` module handles customer complaints, ratings, and service quality tracking.

**Feedback Entities**:

| Entity        | Purpose                                   |
| ------------- | ----------------------------------------- |
| `Complaint`   | Customer issue requiring resolution       |
| `Rating`      | Numeric + optional text feedback          |
| `StaffRating` | Per-staff performance tracking (optional) |

**Complaint Workflow**:

```
Customer files complaint
        │
        ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│    Filed      │────►│  In Progress  │────►│   Resolved    │
│   (new)       │     │  (assigned)   │     │   (closed)    │
└───────────────┘     └───────────────┘     └───────────────┘
        │                                           │
        └─────────► Escalated ◄─────────────────────┘
                   (if unresolved)
```

**Complaint Fields**:

- `id`, `tenant_id`, `order_id` (optional), `customer_id` (optional)
- `type`: food_quality, service, wait_time, billing, cleanliness, other
- `description`: Free text
- `severity`: low, medium, high, critical
- `status`: filed, in_progress, escalated, resolved, closed
- `assigned_to`: Staff member handling
- `resolution`: How it was resolved
- `created_at`, `resolved_at`

**Rating Fields**:

- `id`, `tenant_id`, `order_id`, `customer_id` (optional)
- `overall_rating`: 1-5 stars
- `food_rating`: 1-5 (optional)
- `service_rating`: 1-5 (optional)
- `comment`: Free text (optional)
- `created_at`

**Feedback API**:

```graphql
type Mutation {
  fileComplaint(input: ComplaintInput!): Complaint!
  updateComplaintStatus(id: ID!, status: ComplaintStatus!, resolution: String): Complaint!
  submitRating(input: RatingInput!): Rating!
}

type Query {
  complaints(filter: ComplaintFilter, pagination: Pagination): ComplaintConnection!
  ratings(filter: RatingFilter, pagination: Pagination): RatingConnection!
  feedbackSummary(tenantId: ID!, period: DateRange!): FeedbackSummary!
}
```

**Feedback Analytics**:

| Metric                    | Calculation                            |
| ------------------------- | -------------------------------------- |
| Average rating            | Mean of `overall_rating`               |
| NPS (Net Promoter Score)  | % promoters (4-5) - % detractors (1-2) |
| Complaint resolution time | Avg time from `filed` to `resolved`    |
| Complaints by type        | Aggregation by `type` field            |
| Rating trends             | Ratings over time                      |

**Rationale**: Observability enables debugging and performance analysis. Clickstream enables
data-driven menu and pricing decisions. Feedback closes the loop with customers and tracks
service quality over time.

### XII. Security Architecture

The system MUST implement defense-in-depth security with clear patterns for authentication,
authorization, data protection, and secure communication.

#### Authentication

**JWT-Based Authentication**:

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  Login  │────►│  Auth   │────►│  Issue  │────►│ Access  │
│ Request │     │ Service │     │   JWT   │     │   API   │
└─────────┘     └─────────┘     └─────────┘     └─────────┘
                                     │
                              ┌──────┴──────┐
                              │   Refresh   │
                              │    Token    │
                              └─────────────┘
```

**Token Requirements**:

- Access tokens MUST expire in 15-60 minutes
- Refresh tokens MUST expire in 7-30 days
- Refresh tokens MUST be stored securely (httpOnly cookie or secure storage)
- Token rotation MUST occur on refresh
- Revocation MUST be supported (token blacklist or short expiry + refresh)

**JWT Payload**:

```json
{
  "sub": "user-uuid",
  "tid": "tenant-uuid",
  "role": "manager",
  "permissions": ["orders:write", "inventory:read"],
  "iat": 1706500000,
  "exp": 1706503600
}
```

#### Authorization (RBAC)

**Role Hierarchy**:

| Role      | Description               | Example Permissions                     |
| --------- | ------------------------- | --------------------------------------- |
| `owner`   | Tenant owner, full access | `*:*`                                   |
| `admin`   | Administrative access     | `users:*, config:*, reports:*`          |
| `manager` | Operational management    | `orders:*, inventory:*, staff:read`     |
| `cashier` | Payment and billing       | `orders:read, payments:*, billing:*`    |
| `waiter`  | Order taking              | `orders:write, tables:write, menu:read` |
| `kitchen` | Kitchen operations        | `orders:read, kitchen:write`            |
| `viewer`  | Read-only access          | `*:read`                                |

**Permission Format**: `resource:action` or `resource:action:scope`

- `orders:write` - Can create/update orders
- `orders:read:own` - Can read only own orders
- `reports:read:tenant` - Can read tenant reports

**Authorization Enforcement**:

```go
// Middleware checks permissions before handler
func RequirePermission(permission string) Middleware {
    return func(next Handler) Handler {
        return func(ctx context.Context, req Request) Response {
            user := auth.UserFromContext(ctx)
            if !user.HasPermission(permission) {
                return ForbiddenError("insufficient permissions")
            }
            return next(ctx, req)
        }
    }
}
```

#### API Security

**Rate Limiting**:

| Endpoint Type     | Limit         | Window   | Scope       |
| ----------------- | ------------- | -------- | ----------- |
| Authentication    | 5 requests    | 1 minute | Per IP      |
| GraphQL mutations | 100 requests  | 1 minute | Per user    |
| GraphQL queries   | 500 requests  | 1 minute | Per user    |
| REST integrations | 1000 requests | 1 minute | Per API key |
| File uploads      | 10 requests   | 1 minute | Per user    |

**Input Validation**:

- All inputs MUST be validated at API boundary
- GraphQL inputs MUST use input types with validation directives
- SQL queries MUST use parameterized queries (never string concatenation)
- File uploads MUST validate type, size, and content

**CORS Configuration**:

```go
cors := cors.Config{
    AllowedOrigins:   []string{"https://app.solobueno.com", "capacitor://localhost"},
    AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
    AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Tenant-ID"},
    AllowCredentials: true,
    MaxAge:           86400,
}
```

#### Data Protection

**Encryption at Rest**:

- Database MUST use encrypted storage (PostgreSQL TDE or disk encryption)
- Backups MUST be encrypted
- Sensitive fields MAY use application-level encryption (e.g., PII)

**Encryption in Transit**:

- All external communication MUST use TLS 1.2+
- Internal service communication SHOULD use TLS (MUST in production)
- WebSocket connections MUST use WSS

**Sensitive Data Handling**:

| Data Type          | Storage                | Logging  | Events             |
| ------------------ | ---------------------- | -------- | ------------------ |
| Passwords          | Hashed (bcrypt/argon2) | ❌ Never | ❌ Never           |
| Payment cards      | ❌ Never stored        | ❌ Never | ❌ Never           |
| Tokens             | Hashed or encrypted    | ❌ Never | ❌ Never           |
| PII (email, phone) | Encrypted or plain     | Masked   | Masked or excluded |
| Session IDs        | Plain                  | ✅ OK    | ✅ OK              |

#### Secret Management

**Environment-Based Secrets**:

| Environment  | Secret Source                        | Examples               |
| ------------ | ------------------------------------ | ---------------------- |
| dev          | `.env` file (gitignored)             | Local DB password      |
| test         | CI secrets                           | Test API keys          |
| staging/prod | Vault / AWS SSM / GCP Secret Manager | Production credentials |

**Secret Rotation**:

- Database credentials SHOULD rotate quarterly
- API keys MUST be rotatable without downtime
- JWT signing keys MUST support rotation (kid header)

#### OWASP Top 10 Considerations

| Risk                      | Mitigation                                        |
| ------------------------- | ------------------------------------------------- |
| Injection                 | Parameterized queries, input validation           |
| Broken Authentication     | JWT best practices, rate limiting, MFA (future)   |
| Sensitive Data Exposure   | Encryption, minimal data exposure, secure headers |
| XML External Entities     | N/A (JSON only)                                   |
| Broken Access Control     | RBAC enforcement at every layer, tenant isolation |
| Security Misconfiguration | Secure defaults, environment-specific configs     |
| XSS                       | React auto-escaping, CSP headers                  |
| Insecure Deserialization  | Type-safe JSON parsing, schema validation         |
| Vulnerable Components     | Dependency scanning, regular updates              |
| Insufficient Logging      | Audit trail, security event logging               |

**Security Headers**:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'; script-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
```

**Rationale**: Restaurant ERP handles payments, PII, and multi-tenant data. Security breaches
can cause financial loss, legal liability, and reputation damage. Defense-in-depth ensures
no single failure compromises the system.

## Technology & Constraints

### Monorepo Structure

```
solobueno-erp/
├── apps/
│   ├── mobile/                 # React Native app (orders, tables)
│   ├── kitchen-display/        # React Native tablet app
│   ├── backoffice/             # React web app
│   └── admin/                  # React web app
│
├── backend/
│   ├── cmd/
│   │   ├── server/             # Main application entrypoint
│   │   └── migrate/            # Migration CLI tool
│   ├── internal/
│   │   ├── auth/               # Auth module
│   │   ├── menu/               # Menu module
│   │   ├── orders/             # Orders module
│   │   ├── tables/             # Tables module
│   │   ├── inventory/          # Inventory module
│   │   ├── billing/            # Billing module
│   │   ├── payments/           # Payments module
│   │   ├── reporting/          # Reporting module
│   │   ├── config/             # Config module
│   │   ├── feedback/           # Complaints & ratings module
│   │   ├── analytics/          # Clickstream & BI events
│   │   ├── notifications/      # Push, email, SMS, in-app
│   │   ├── audit/              # Audit trail & compliance logging
│   │   ├── jobs/               # Background jobs & scheduling
│   │   ├── media/              # File storage & image processing
│   │   ├── search/             # Full-text search (optional)
│   │   └── shared/             # Shared kernel (minimal)
│   │       ├── events/         # Event bus, base event types
│   │       ├── types/          # TenantID, Money, etc.
│   │       ├── saga/           # Saga orchestration
│   │       ├── observability/  # Logger, metrics, tracer interfaces
│   │       ├── cache/          # Cache adapter interface
│   │       ├── errors/         # Error types & handling
│   │       └── resilience/     # Circuit breaker, retry policies
│   ├── api/
│   │   ├── graphql/            # GraphQL schema & resolvers
│   │   └── rest/               # REST handlers (integrations)
│   ├── migrations/             # Database migrations
│   │   └── ...
│   └── plugins/                # Plugin implementations
│       ├── billing/
│       │   ├── costarica/      # Costa Rica Hacienda FE
│       │   └── generic/        # Generic invoicing
│       ├── payments/
│       │   ├── stripe/         # Stripe integration
│       │   └── manual/         # Cash/manual payments
│       └── analytics/          # Analytics export plugins
│           └── mixpanel/       # Mixpanel integration (optional)
│
├── packages/
│   ├── ui/                     # Shared React/RN components (@solobueno/ui)
│   ├── i18n/                   # Translations package (@solobueno/i18n)
│   ├── types/                  # Shared TypeScript types (@solobueno/types)
│   ├── graphql-client/         # Generated GraphQL client (@solobueno/graphql)
│   └── analytics/              # Client-side analytics helpers (@solobueno/analytics)
│
├── tools/
│   ├── codegen/                # GraphQL code generator config
│   └── scripts/                # Build, deploy scripts
│
├── docs/
│   ├── adr/                    # Architecture Decision Records
│   ├── events/                 # Event catalog documentation
│   └── api/                    # API documentation
│
└── infrastructure/
    ├── docker/
    │   ├── Dockerfile.backend
    │   ├── Dockerfile.mobile   # For CI builds
    │   └── docker-compose.yml  # Local development
    ├── k8s/                    # Kubernetes manifests (if used)
    │   ├── base/
    │   └── overlays/
    │       ├── dev/
    │       ├── test/
    │       ├── staging/
    │       └── prod/
    ├── terraform/              # Infrastructure as Code (if used)
    │   ├── modules/
    │   └── environments/
    └── config/
        ├── dev.env.example
        ├── test.env.example
        ├── staging.env.example
        └── prod.env.example
```

### Stack Requirements (Finalized)

| Layer                  | Technology                | Rationale                                             |
| ---------------------- | ------------------------- | ----------------------------------------------------- |
| **Mobile App**         | React Native + TypeScript | Cross-platform, type safety, code sharing             |
| **Web Apps**           | React + TypeScript        | Shared components with mobile                         |
| **Backend**            | Go 1.22+                  | Performance, strong typing, excellent concurrency     |
| **GraphQL**            | gqlgen                    | Native Go, type-safe resolvers, subscriptions         |
| **REST**               | Chi                       | Lightweight, idiomatic Go, good middleware            |
| **Database**           | PostgreSQL 16             | Multi-tenant, JSONB flexibility, reliability          |
| **Migrations**         | golang-migrate            | Widely used, CLI + library, SQL-based                 |
| **Cache**              | Redis 7                   | Shared cache, session storage, job queues             |
| **Logging**            | zerolog                   | Structured, zero-allocation, high-performance         |
| **Event Bus (Dev)**    | In-memory                 | Simple, fast, good for development                    |
| **Event Bus (Prod)**   | NATS                      | Purpose-built for events, persistence, replay         |
| **Job Queue**          | Asynq                     | Redis-based, well-maintained, monitoring UI           |
| **File Storage**       | AWS S3                    | Lightsail integration, CDN-compatible, cost-effective |
| **Search**             | PostgreSQL FTS            | Start simple; Meilisearch if needed later             |
| **Push Notifications** | FCM + APNs                | Industry standard for mobile push                     |
| **Email**              | AWS SES                   | Same ecosystem as Lightsail, cost-effective           |
| **Offline Sync**       | WatermelonDB              | React Native optimized, excellent performance         |
| **Real-time**          | GraphQL Subscriptions     | Backed by domain events via NATS                      |
| **Monorepo Tool**      | Turborepo                 | Fast, simple config, good caching                     |
| **CI/CD**              | GitHub Actions            | GitHub integration, generous free tier                |
| **Container Registry** | Amazon ECR                | AWS ecosystem, Lightsail integration                  |
| **Secrets**            | AWS SSM Parameter Store   | AWS ecosystem, free tier, encryption                  |
| **Cloud Provider**     | AWS Lightsail             | Cost-effective, predictable pricing, upgrade path     |
| **Primary Region**     | us-east-1                 | Closest to Costa Rica with full Lightsail support     |

### Module Design Pattern

Each domain module follows this internal structure:

```
internal/orders/
├── domain/                 # Domain types, entities, value objects
│   ├── order.go
│   ├── cart.go
│   └── events.go           # OrderCreated, OrderCompleted, etc.
├── ports/                  # Interfaces (driven & driving)
│   ├── repository.go       # Data access interface
│   ├── service.go          # Public module API interface
│   └── events.go           # Event publisher interface
├── adapters/               # Interface implementations
│   ├── postgres/           # PostgreSQL repository
│   └── memory/             # In-memory (for testing)
├── application/            # Use cases / application services
│   ├── service.go          # Command handlers
│   ├── queries.go          # Query handlers
│   └── subscribers.go      # Event handlers from other modules
└── module.go               # Module initialization, DI wiring, event subscriptions
```

### Database Architecture

#### PostgreSQL as Default

PostgreSQL MUST be the default database for all modules. It handles:

- Transactional data (orders, payments, inventory)
- Time-series data (analytics events with partitioning)
- JSON flexibility (event payloads, configuration)
- Full-text search (menu items, complaints)

#### Per-Module Database Isolation

Each module MUST own its data exclusively, enabling independent database decisions:

**Schema-Per-Module Pattern** (recommended for monolith):

```
PostgreSQL Database: solobueno_erp
├── schema: auth          # auth module tables
├── schema: menu          # menu module tables
├── schema: orders        # orders module tables
├── schema: tables        # tables module tables
├── schema: inventory     # inventory module tables
├── schema: billing       # billing module tables
├── schema: payments      # payments module tables
├── schema: reporting     # reporting module tables
├── schema: config        # config module tables
├── schema: i18n          # i18n module tables
├── schema: feedback      # feedback module tables
├── schema: analytics     # analytics module tables (partitioned)
└── schema: shared        # shared lookup tables (minimal)
```

```sql
-- Example: Each module has its own schema
CREATE SCHEMA IF NOT EXISTS orders;
CREATE SCHEMA IF NOT EXISTS inventory;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Tables live in their module's schema
CREATE TABLE orders.orders (...);
CREATE TABLE orders.order_items (...);
CREATE TABLE inventory.stock_levels (...);
CREATE TABLE analytics.events (...);

-- Cross-module references use schema-qualified names
-- But prefer ID references over foreign keys across schemas
```

#### Repository Adapter Pattern

Each module defines a repository interface (port) and can have multiple implementations (adapters):

```go
// orders/ports/repository.go - THE CONTRACT
type OrderRepository interface {
    Save(ctx context.Context, order *domain.Order) error
    FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
    FindByTenant(ctx context.Context, tenantID TenantID, filter OrderFilter) ([]*domain.Order, error)
    Delete(ctx context.Context, id domain.OrderID) error
}

// orders/adapters/postgres/repository.go - PostgreSQL IMPLEMENTATION
type PostgresOrderRepository struct {
    db     *sql.DB
    schema string  // "orders"
}

func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
    query := fmt.Sprintf(`INSERT INTO %s.orders ...`, r.schema)
    // ... implementation
}

// orders/adapters/memory/repository.go - In-memory IMPLEMENTATION (for tests)
type MemoryOrderRepository struct {
    orders map[domain.OrderID]*domain.Order
    mu     sync.RWMutex
}
```

#### Swapping Database Per Module

The adapter pattern allows any module to switch databases without affecting others:

```go
// Module initialization - database choice is configuration
func NewOrdersModule(cfg Config) *OrdersModule {
    var repo ports.OrderRepository

    switch cfg.Orders.DatabaseType {
    case "postgres":
        repo = postgres.NewOrderRepository(cfg.Orders.PostgresURL, "orders")
    case "memory":
        repo = memory.NewOrderRepository()  // For testing
    // Future: could add other implementations
    // case "cockroachdb":
    //     repo = cockroach.NewOrderRepository(cfg.Orders.CockroachURL)
    default:
        repo = postgres.NewOrderRepository(cfg.Orders.PostgresURL, "orders")
    }

    return &OrdersModule{
        repo:    repo,
        service: application.NewOrderService(repo, ...),
    }
}
```

#### Analytics Module - PostgreSQL with Partitioning

The analytics module uses PostgreSQL with time-based partitioning for high-volume events:

```sql
-- analytics schema with partitioned events table
CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE analytics.events (
    id              UUID DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    session_id      UUID NOT NULL,
    user_id         UUID,
    payload         JSONB NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, occurred_at, id)
) PARTITION BY RANGE (occurred_at);

-- Monthly partitions
CREATE TABLE analytics.events_2025_01 PARTITION OF analytics.events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE analytics.events_2025_02 PARTITION OF analytics.events
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Pre-aggregated summaries for fast queries
CREATE TABLE analytics.daily_summary (
    tenant_id       UUID NOT NULL,
    date            DATE NOT NULL,
    metrics         JSONB NOT NULL,  -- Flexible aggregations
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, date)
);
```

#### Database Connection Configuration

```go
type DatabaseConfig struct {
    // Shared PostgreSQL connection (default)
    DefaultPostgresURL string

    // Per-module overrides (optional)
    Modules map[string]ModuleDatabaseConfig
}

type ModuleDatabaseConfig struct {
    Type     string // "postgres" (default), "memory" (testing)
    URL      string // Connection string (if different from default)
    Schema   string // Schema name within PostgreSQL
    PoolSize int    // Connection pool size
}

// Example configuration
config := DatabaseConfig{
    DefaultPostgresURL: "postgres://localhost/solobueno_erp",
    Modules: map[string]ModuleDatabaseConfig{
        "orders":    {Schema: "orders", PoolSize: 20},
        "analytics": {Schema: "analytics", PoolSize: 10},
        "feedback":  {Schema: "feedback", PoolSize: 5},
        // All use same PostgreSQL, different schemas
    },
}
```

#### Cross-Module Data Access Rules

- Modules MUST NOT directly query another module's tables
- Cross-module data access MUST go through the module's service interface
- Foreign key references across schemas SHOULD be avoided (use soft references by ID)
- Shared lookup data (countries, currencies) MAY live in a `shared` schema

```go
// ❌ WRONG: Orders module directly queries inventory table
func (s *OrderService) CreateOrder(...) {
    row := s.db.QueryRow("SELECT quantity FROM inventory.stock WHERE item_id = $1", itemID)
}

// ✅ CORRECT: Orders module calls inventory service interface
func (s *OrderService) CreateOrder(...) {
    available, err := s.inventory.CheckStock(ctx, itemID, quantity)
    if !available {
        return ErrInsufficientStock
    }
}
```

#### Migration Organization

Migrations are organized by module schema:

```
backend/migrations/
├── 000001_create_schemas.up.sql           # Create all schemas
├── 000001_create_schemas.down.sql
├── 000002_auth_create_tables.up.sql       # auth schema tables
├── 000002_auth_create_tables.down.sql
├── 000003_menu_create_tables.up.sql       # menu schema tables
├── 000003_menu_create_tables.down.sql
├── 000004_orders_create_tables.up.sql     # orders schema tables
├── 000004_orders_create_tables.down.sql
├── ...
├── 0000XX_analytics_create_events.up.sql  # analytics partitioned table
├── 0000XX_analytics_create_events.down.sql
├── 0000XX_feedback_create_tables.up.sql   # feedback schema tables
└── 0000XX_feedback_create_tables.down.sql
```

```sql
-- 000001_create_schemas.up.sql
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS menu;
CREATE SCHEMA IF NOT EXISTS orders;
CREATE SCHEMA IF NOT EXISTS tables;
CREATE SCHEMA IF NOT EXISTS inventory;
CREATE SCHEMA IF NOT EXISTS billing;
CREATE SCHEMA IF NOT EXISTS payments;
CREATE SCHEMA IF NOT EXISTS reporting;
CREATE SCHEMA IF NOT EXISTS config;
CREATE SCHEMA IF NOT EXISTS i18n;
CREATE SCHEMA IF NOT EXISTS feedback;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS shared;

-- 000001_create_schemas.down.sql
DROP SCHEMA IF EXISTS analytics CASCADE;
DROP SCHEMA IF EXISTS feedback CASCADE;
-- ... etc (reverse order)
```

**Rationale**: PostgreSQL handles all current requirements (transactional, time-series, JSON, search).
Schema-per-module provides isolation while keeping operational simplicity. Repository adapters ensure
any module can migrate to a different database if specialized needs arise (e.g., ClickHouse for
analytics at extreme scale) without affecting other modules.

### Caching Strategy

Caching MUST be implemented to optimize performance and support offline-first operations.

#### Cache Layers

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Client (Mobile/Web)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │
│  │   Memory    │  │   SQLite/   │  │   Assets    │                  │
│  │   Cache     │  │ WatermelonDB│  │    Cache    │                  │
│  └─────────────┘  └─────────────┘  └─────────────┘                  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                            CDN Layer                                 │
│              (Static assets, menu images, i18n bundles)              │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Backend API                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │
│  │  In-Memory  │  │    Redis    │  │  PostgreSQL │                  │
│  │ (per-request)│  │ (shared)    │  │   (source)  │                  │
│  └─────────────┘  └─────────────┘  └─────────────┘                  │
└─────────────────────────────────────────────────────────────────────┘
```

#### What to Cache

| Data               | Cache Location | TTL              | Invalidation             |
| ------------------ | -------------- | ---------------- | ------------------------ |
| Menu items         | Redis + Client | 1 hour           | On menu update event     |
| Menu categories    | Redis + Client | 1 hour           | On category update event |
| Tenant config      | Redis + Client | 5 minutes        | On config change         |
| Feature flags      | Redis + Client | 1 minute         | On flag toggle           |
| Translations       | Redis + Client | 1 hour           | On i18n update           |
| User session       | Redis          | Session lifetime | On logout                |
| Table layout       | Redis + Client | 30 minutes       | On layout change         |
| Current orders     | Client only    | Real-time        | WebSocket updates        |
| Price calculations | Per-request    | Request only     | N/A                      |

#### Cache Adapter Pattern

```go
// Cache interface - implementations can be swapped
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    DeletePattern(ctx context.Context, pattern string) error
}

// Implementations
├── memory/cache.go     # In-memory (dev, per-instance)
├── redis/cache.go      # Redis (production, shared)
└── noop/cache.go       # No-op (testing)
```

#### Cache Key Patterns

```
tenant:{tenant_id}:menu:items           # All menu items for tenant
tenant:{tenant_id}:menu:item:{item_id}  # Single menu item
tenant:{tenant_id}:config               # Tenant configuration
tenant:{tenant_id}:i18n:{locale}        # Translations
user:{user_id}:session                  # User session data
```

#### Cache Invalidation

Event-driven invalidation ensures consistency:

```go
// Subscribe to menu events and invalidate cache
func (c *MenuCache) OnMenuItemUpdated(ctx context.Context, event MenuItemUpdated) error {
    keys := []string{
        fmt.Sprintf("tenant:%s:menu:items", event.TenantID),
        fmt.Sprintf("tenant:%s:menu:item:%s", event.TenantID, event.ItemID),
    }
    return c.cache.Delete(ctx, keys...)
}
```

### Error Handling & Resilience

The system MUST handle errors gracefully with consistent patterns and resilience mechanisms.

#### Error Types

```go
// Base error types
type AppError struct {
    Code       string         // Machine-readable code: "ORDER_NOT_FOUND"
    Message    string         // Human-readable message
    HTTPStatus int            // HTTP status code
    Details    map[string]any // Additional context
    Cause      error          // Underlying error
}

// Error categories
const (
    ErrValidation   = "VALIDATION_ERROR"    // 400 - Bad input
    ErrUnauthorized = "UNAUTHORIZED"        // 401 - Not authenticated
    ErrForbidden    = "FORBIDDEN"           // 403 - Not authorized
    ErrNotFound     = "NOT_FOUND"           // 404 - Resource not found
    ErrConflict     = "CONFLICT"            // 409 - State conflict
    ErrInternal     = "INTERNAL_ERROR"      // 500 - Server error
    ErrUnavailable  = "SERVICE_UNAVAILABLE" // 503 - Temporary failure
)
```

#### GraphQL Error Format

```json
{
  "errors": [
    {
      "message": "Order not found",
      "extensions": {
        "code": "ORDER_NOT_FOUND",
        "details": {
          "order_id": "uuid-here"
        }
      },
      "path": ["order", "byId"]
    }
  ]
}
```

#### Circuit Breaker Pattern

For external services (payment providers, fiscal APIs):

```go
type CircuitBreaker struct {
    name          string
    maxFailures   int           // Failures before opening
    timeout       time.Duration // Time before half-open
    state         State         // closed, open, half-open
}

// Usage
paymentBreaker := NewCircuitBreaker("stripe", 5, 30*time.Second)

func (s *PaymentService) ProcessPayment(ctx context.Context, req PaymentRequest) error {
    return s.breaker.Execute(func() error {
        return s.stripeClient.Charge(ctx, req)
    })
}
```

**Circuit States**:

- **Closed**: Normal operation, requests pass through
- **Open**: Failures exceeded threshold, requests fail fast
- **Half-Open**: After timeout, allow one request to test recovery

#### Retry Policy

```go
type RetryConfig struct {
    MaxAttempts int
    InitialWait time.Duration
    MaxWait     time.Duration
    Multiplier  float64       // Exponential backoff
    Retryable   func(error) bool
}

// Default for external APIs
var DefaultRetry = RetryConfig{
    MaxAttempts: 3,
    InitialWait: 100 * time.Millisecond,
    MaxWait:     5 * time.Second,
    Multiplier:  2.0,
    Retryable:   IsTransientError,
}
```

#### Graceful Degradation

| Failure                | Degraded Behavior                      |
| ---------------------- | -------------------------------------- |
| Payment provider down  | Queue payment, notify user, allow cash |
| Fiscal API unavailable | Queue invoice, generate later          |
| Analytics service slow | Drop analytics events (non-critical)   |
| Search unavailable     | Fall back to basic DB query            |
| Cache unavailable      | Direct database queries (slower)       |

### Background Jobs & Scheduling

The system MUST support asynchronous processing for long-running and scheduled tasks.

#### Job Queue Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Producer  │────►│    Queue    │────►│   Worker    │
│  (API/Event)│     │  (Redis/PG) │     │  (Consumer) │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │   Dead      │
                    │   Letter    │
                    │   Queue     │
                    └─────────────┘
```

#### Job Types

| Job                      | Trigger                  | Priority | Retry |
| ------------------------ | ------------------------ | -------- | ----- |
| Generate invoice PDF     | `PaymentSucceeded` event | High     | 3x    |
| Send email notification  | `InvoiceGenerated` event | Medium   | 5x    |
| Process analytics batch  | Scheduled (5 min)        | Low      | 3x    |
| Generate daily report    | Scheduled (daily 3am)    | Low      | 3x    |
| Cleanup expired sessions | Scheduled (hourly)       | Low      | 1x    |
| Create partition tables  | Scheduled (monthly)      | High     | 3x    |
| Sync offline data        | `DeviceOnline` event     | High     | 5x    |
| Process image upload     | `FileUploaded` event     | Medium   | 3x    |

#### Job Definition

```go
type Job struct {
    ID          string
    Type        string
    TenantID    string
    Payload     json.RawMessage
    Priority    Priority        // high, medium, low
    ScheduledAt time.Time       // When to run (for scheduled jobs)
    MaxRetries  int
    Attempts    int
    Status      JobStatus       // pending, running, completed, failed
    Error       string
    CreatedAt   time.Time
    StartedAt   *time.Time
    CompletedAt *time.Time
}

// Job handler interface
type JobHandler interface {
    Handle(ctx context.Context, job *Job) error
}
```

#### Scheduled Jobs

```go
// Cron-like scheduling
scheduler.Register("analytics:aggregate", "*/5 * * * *", analyticsHandler)  // Every 5 min
scheduler.Register("reports:daily", "0 3 * * *", dailyReportHandler)        // Daily 3am
scheduler.Register("cleanup:sessions", "0 * * * *", sessionCleanupHandler)  // Hourly
scheduler.Register("partitions:create", "0 0 1 * *", partitionHandler)      // Monthly
```

#### Job Monitoring

- Jobs MUST emit `JobStarted`, `JobCompleted`, `JobFailed` events
- Failed jobs MUST be logged with full context
- Dead letter queue MUST be monitored and alerted
- Job metrics: queue depth, processing time, failure rate

### File Storage & Media

The system MUST handle file uploads with proper validation, processing, and storage.

#### Storage Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Upload    │────►│  Validate   │────►│   Process   │────►│   Store     │
│   Request   │     │  (type/size)│     │  (resize)   │     │  (S3/local) │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                            ┌──────┴──────┐
                                                            │     CDN     │
                                                            └─────────────┘
```

#### Storage Adapter Pattern

```go
type FileStorage interface {
    Upload(ctx context.Context, file File, path string) (URL, error)
    Download(ctx context.Context, path string) (io.Reader, error)
    Delete(ctx context.Context, path string) error
    GenerateSignedURL(ctx context.Context, path string, ttl time.Duration) (URL, error)
}

// Implementations
├── local/storage.go    # Local filesystem (dev)
├── s3/storage.go       # AWS S3 / MinIO
├── gcs/storage.go      # Google Cloud Storage
└── memory/storage.go   # In-memory (testing)
```

#### File Types & Limits

| Type         | Allowed Formats | Max Size | Processing                    |
| ------------ | --------------- | -------- | ----------------------------- |
| Menu images  | JPEG, PNG, WebP | 5 MB     | Resize to 800x800, thumbnails |
| Receipt PDF  | PDF             | 2 MB     | None                          |
| Invoice PDF  | PDF             | 2 MB     | None (generated)              |
| User avatar  | JPEG, PNG       | 1 MB     | Resize to 200x200             |
| Import files | CSV, XLSX       | 10 MB    | Validation                    |

#### Image Processing

```go
type ImageProcessor interface {
    Resize(img image.Image, width, height int) image.Image
    GenerateThumbnail(img image.Image, size int) image.Image
    ConvertToWebP(img image.Image) ([]byte, error)
}

// Menu item image variants
variants := []ImageVariant{
    {Name: "original", MaxWidth: 1200, MaxHeight: 1200},
    {Name: "display", MaxWidth: 800, MaxHeight: 800},
    {Name: "thumbnail", MaxWidth: 200, MaxHeight: 200},
}
```

#### Storage Path Convention

```
{tenant_id}/
├── menu/
│   └── {item_id}/
│       ├── original.webp
│       ├── display.webp
│       └── thumbnail.webp
├── invoices/
│   └── {year}/{month}/
│       └── {invoice_id}.pdf
├── avatars/
│   └── {user_id}.webp
└── imports/
    └── {import_id}/
        └── source.csv
```

### Notifications

The system MUST support multi-channel notifications with delivery tracking.

#### Notification Channels

| Channel         | Use Cases                         | Provider       |
| --------------- | --------------------------------- | -------------- |
| Push (Mobile)   | Order ready, payment received     | FCM / APNs     |
| Email           | Invoices, reports, password reset | SendGrid / SES |
| SMS             | Critical alerts, 2FA (future)     | Twilio / SNS   |
| In-App          | All notifications                 | WebSocket      |
| Kitchen Display | New orders, modifications         | WebSocket      |

#### Notification Architecture

```go
type Notification struct {
    ID          string
    TenantID    string
    UserID      string            // Target user (or nil for broadcast)
    Channel     Channel           // push, email, sms, in_app
    Template    string            // Template identifier
    Data        map[string]any    // Template variables
    Priority    Priority          // high, medium, low
    Status      Status            // pending, sent, delivered, failed
    ScheduledAt *time.Time        // For scheduled notifications
    SentAt      *time.Time
    ReadAt      *time.Time
}

type NotificationService interface {
    Send(ctx context.Context, notification Notification) error
    SendBatch(ctx context.Context, notifications []Notification) error
    MarkAsRead(ctx context.Context, id string, userID string) error
    GetUnread(ctx context.Context, userID string) ([]Notification, error)
}
```

#### Notification Templates

```go
// Template-based notifications
templates := map[string]Template{
    "order_ready": {
        Push:  "Your order #{{.OrderNumber}} is ready for pickup!",
        Email: "order_ready.html",
        SMS:   "Order #{{.OrderNumber}} ready at {{.RestaurantName}}",
    },
    "payment_received": {
        Push:  "Payment of {{.Amount}} received. Thank you!",
        Email: "payment_received.html",
    },
    "low_stock_alert": {
        Push:  "Low stock alert: {{.ItemName}} ({{.Quantity}} remaining)",
        Email: "low_stock.html",
    },
}
```

#### User Preferences

```go
type NotificationPreferences struct {
    UserID    string
    TenantID  string
    Channels  map[string]bool    // {"push": true, "email": true, "sms": false}
    Quiet     *QuietHours        // No notifications during these hours
    Frequency string             // "immediate", "hourly_digest", "daily_digest"
}
```

### Real-Time Architecture

The system MUST support real-time updates for order tracking, kitchen display, and live dashboards.

#### WebSocket / Subscription Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        GraphQL Subscriptions                         │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Subscription Manager                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │
│  │  Connection │  │   Topic     │  │   Message   │                  │
│  │   Registry  │  │   Router    │  │   Fanout    │                  │
│  └─────────────┘  └─────────────┘  └─────────────┘                  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Event Bus                                    │
│              (Domain events trigger subscription updates)            │
└─────────────────────────────────────────────────────────────────────┘
```

#### Subscription Topics

```graphql
type Subscription {
  # Order updates for a specific table/session
  orderUpdated(tableId: ID!): Order!

  # Kitchen display - new orders to prepare
  kitchenOrders(tenantId: ID!): KitchenOrder!

  # Order status changes
  orderStatusChanged(orderId: ID!): OrderStatus!

  # Real-time inventory alerts
  inventoryAlert(tenantId: ID!): InventoryAlert!

  # Live dashboard metrics
  dashboardMetrics(tenantId: ID!): DashboardMetrics!
}
```

#### Connection Management

```go
type ConnectionManager struct {
    connections map[string]*Connection  // connectionID -> Connection
    topics      map[string][]string     // topic -> []connectionID
    mu          sync.RWMutex
}

type Connection struct {
    ID        string
    UserID    string
    TenantID  string
    Topics    []string
    CreatedAt time.Time
    LastPing  time.Time
}

// Heartbeat to detect stale connections
func (m *ConnectionManager) StartHeartbeat() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        m.removeStaleConnections(60 * time.Second)
    }
}
```

#### Event-to-Subscription Bridge

```go
// Domain events automatically trigger subscription updates
func (b *SubscriptionBridge) OnOrderStatusChanged(ctx context.Context, event OrderStatusChanged) {
    b.subscriptions.Publish(ctx, Subscription{
        Topic:    fmt.Sprintf("order:%s:status", event.OrderID),
        TenantID: event.TenantID,
        Payload:  event,
    })
}
```

### API Versioning

The system MUST support API evolution without breaking existing clients.

#### GraphQL Schema Evolution

**Additive Changes** (non-breaking):

- Add new fields to types
- Add new types
- Add new queries/mutations
- Add new enum values (at end)
- Add optional arguments

**Breaking Changes** (require deprecation):

- Remove fields
- Change field types
- Remove enum values
- Change argument types
- Remove queries/mutations

#### Deprecation Strategy

```graphql
type Order {
  id: ID!
  status: OrderStatus!

  # Deprecated field - use 'status' instead
  orderStatus: String @deprecated(reason: "Use 'status' field instead. Removal: 2025-06-01")

  # New field
  statusHistory: [StatusChange!]!
}
```

**Deprecation Timeline**:

1. Add deprecation warning (GraphQL `@deprecated` directive)
2. Log usage of deprecated fields
3. Notify clients using deprecated fields
4. Minimum 3-month deprecation period
5. Remove in next major version

#### REST API Versioning

```
# URL-based versioning
/api/v1/orders
/api/v2/orders

# Header-based versioning (alternative)
X-API-Version: 2
```

**Version Support Policy**:

- Support current version (N) and previous version (N-1)
- Deprecation notice 6 months before removal
- Security fixes backported to supported versions

#### Mobile App Compatibility

```go
// Minimum supported version check
type VersionCheck struct {
    MinVersion     string  // Minimum app version required
    CurrentVersion string  // Latest app version
    ForceUpdate    bool    // Block access if below minimum
    UpdateURL      string  // App store URL
}

// API response for outdated clients
{
    "error": "CLIENT_VERSION_OUTDATED",
    "minVersion": "2.0.0",
    "currentVersion": "2.5.0",
    "forceUpdate": true,
    "updateUrl": "https://apps.apple.com/..."
}
```

### Audit Trail

The system MUST maintain comprehensive audit logs for compliance and debugging.

#### Audit Event Structure

```go
type AuditEvent struct {
    ID          string
    TenantID    string
    ActorID     string           // User who performed action
    ActorType   string           // user, system, api_key
    Action      string           // create, update, delete, access
    Resource    string           // orders, payments, users
    ResourceID  string           // ID of affected resource
    Changes     json.RawMessage  // Before/after for updates
    Metadata    map[string]any   // Additional context
    IPAddress   string
    UserAgent   string
    Timestamp   time.Time
}
```

#### What to Audit

| Category            | Actions                                         | Retention |
| ------------------- | ----------------------------------------------- | --------- |
| Authentication      | Login, logout, password change, failed attempts | 1 year    |
| Authorization       | Permission changes, role assignments            | 1 year    |
| Financial           | Orders, payments, refunds, invoices             | 7 years   |
| Data Access         | Export, bulk read of sensitive data             | 1 year    |
| Configuration       | Setting changes, feature flags                  | 1 year    |
| User Management     | Create, update, delete users                    | 1 year    |
| Critical Operations | Delete operations, bulk updates                 | 1 year    |

#### Audit Storage

```sql
-- Partitioned audit table for efficient querying and retention
CREATE TABLE audit.events (
    id              UUID DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    actor_id        UUID,
    actor_type      VARCHAR(20) NOT NULL,
    action          VARCHAR(50) NOT NULL,
    resource        VARCHAR(100) NOT NULL,
    resource_id     VARCHAR(255),
    changes         JSONB,
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, timestamp, id)
) PARTITION BY RANGE (timestamp);

-- Monthly partitions with automatic cleanup
CREATE TABLE audit.events_2025_01 PARTITION OF audit.events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

#### Audit Query API

```graphql
type Query {
  auditEvents(filter: AuditFilter!, pagination: Pagination): AuditEventConnection!
}

input AuditFilter {
  tenantId: ID!
  actorId: ID
  action: String
  resource: String
  resourceId: String
  dateRange: DateRange!
}
```

### Data Privacy & Compliance

The system MUST handle personal data responsibly and support privacy regulations.

#### PII Identification

| Field         | Classification | Handling                          |
| ------------- | -------------- | --------------------------------- |
| Email         | PII            | Encrypted at rest, masked in logs |
| Phone         | PII            | Encrypted at rest, masked in logs |
| Full name     | PII            | Plain text, excluded from logs    |
| Address       | PII            | Encrypted at rest                 |
| Payment cards | Sensitive      | Never stored (tokenized)          |
| IP address    | PII            | Hashed after 30 days              |
| Device ID     | PII            | Hashed                            |

#### Data Subject Rights (GDPR-aligned)

**Right to Access**:

```go
type DataExportService interface {
    // Export all data for a user in portable format
    ExportUserData(ctx context.Context, userID string) (*DataExport, error)
}

type DataExport struct {
    User          UserData
    Orders        []OrderData
    Payments      []PaymentData
    Preferences   PreferencesData
    AuditHistory  []AuditData
    ExportedAt    time.Time
    Format        string  // JSON, CSV
}
```

**Right to Deletion**:

```go
type DataDeletionService interface {
    // Delete user data (with legal hold check)
    RequestDeletion(ctx context.Context, userID string) (*DeletionRequest, error)

    // Execute deletion after retention period
    ExecuteDeletion(ctx context.Context, requestID string) error
}

// Deletion MUST:
// - Check for legal holds (active disputes, tax requirements)
// - Anonymize rather than delete where legally required
// - Remove from backups (or document retention exception)
// - Log the deletion action (audit trail)
```

**Right to Rectification**:

- Users MUST be able to update their personal data
- Updates MUST be propagated to all copies (cache invalidation)

#### Consent Management

```go
type Consent struct {
    UserID      string
    TenantID    string
    Type        string    // marketing, analytics, cookies
    Granted     bool
    GrantedAt   *time.Time
    RevokedAt   *time.Time
    Source      string    // signup, settings, popup
    Version     string    // Privacy policy version
}
```

#### Data Retention Policies

| Data Type          | Retention  | After Retention        |
| ------------------ | ---------- | ---------------------- |
| Active user data   | Indefinite | Until deletion request |
| Inactive user data | 2 years    | Anonymize or delete    |
| Order history      | 7 years    | Required for tax       |
| Payment records    | 7 years    | Required for tax       |
| Audit logs         | 1-7 years  | Delete                 |
| Analytics events   | 2 years    | Aggregate and delete   |
| Session data       | 30 days    | Delete                 |

### Disaster Recovery

The system MUST have documented backup and recovery procedures.

#### Backup Strategy

| Data              | Frequency        | Retention  | Method                                  |
| ----------------- | ---------------- | ---------- | --------------------------------------- |
| PostgreSQL        | Continuous (WAL) | 30 days    | Streaming replication + daily snapshots |
| PostgreSQL (full) | Daily            | 90 days    | pg_dump to encrypted S3                 |
| Redis             | Hourly           | 7 days     | RDB snapshots                           |
| File storage      | Real-time        | Indefinite | Cross-region replication                |
| Configuration     | On change        | 1 year     | Version control                         |

#### Recovery Objectives

| Metric                             | Target    | Notes                        |
| ---------------------------------- | --------- | ---------------------------- |
| **RPO** (Recovery Point Objective) | < 1 hour  | Maximum data loss acceptable |
| **RTO** (Recovery Time Objective)  | < 4 hours | Maximum downtime acceptable  |

#### Failover Architecture

```
Primary Region                    Secondary Region
┌─────────────────┐              ┌─────────────────┐
│   PostgreSQL    │──streaming──►│   PostgreSQL    │
│   (primary)     │   replication│   (standby)     │
└─────────────────┘              └─────────────────┘

┌─────────────────┐              ┌─────────────────┐
│   Redis         │──────────────│   Redis         │
│   (primary)     │  replication │   (replica)     │
└─────────────────┘              └─────────────────┘

┌─────────────────┐              ┌─────────────────┐
│   App Servers   │              │   App Servers   │
│   (active)      │              │   (standby)     │
└─────────────────┘              └─────────────────┘
```

#### Recovery Procedures

**Database Recovery**:

```bash
# Point-in-time recovery
pg_restore --target-time="2025-01-29 10:30:00" \
           --target-action=promote \
           /backups/basebackup

# Or restore from snapshot
aws rds restore-db-instance-to-point-in-time \
    --source-db-instance-identifier solobueno-prod \
    --target-db-instance-identifier solobueno-recovery \
    --restore-time 2025-01-29T10:30:00Z
```

**Recovery Testing**:

- Full recovery test MUST be performed quarterly
- Recovery procedures MUST be documented in runbooks
- Recovery time MUST be measured and tracked

#### Incident Response

1. **Detection**: Monitoring alerts, user reports
2. **Assessment**: Scope, impact, cause
3. **Containment**: Stop bleeding, preserve evidence
4. **Communication**: Status page, affected tenants
5. **Recovery**: Execute recovery procedures
6. **Post-mortem**: Document and improve

### Search Strategy

The system SHOULD support efficient search for menu items, orders, and other entities.

#### Search Architecture

**Phase 1 - PostgreSQL Full-Text Search** (recommended start):

```sql
-- Add search vectors to menu items
ALTER TABLE menu.items ADD COLUMN search_vector tsvector;

CREATE INDEX idx_menu_items_search ON menu.items USING GIN(search_vector);

-- Update search vector on insert/update
CREATE OR REPLACE FUNCTION menu.update_search_vector() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('spanish', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('spanish', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('spanish', COALESCE(NEW.category_name, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Search query
SELECT * FROM menu.items
WHERE tenant_id = $1
  AND search_vector @@ plainto_tsquery('spanish', $2)
ORDER BY ts_rank(search_vector, plainto_tsquery('spanish', $2)) DESC
LIMIT 20;
```

**Phase 2 - Dedicated Search Service** (if needed):

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Database   │────►│   Indexer   │────►│   Search    │
│  (source)   │     │   (sync)    │     │   Engine    │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
                                        Elasticsearch
                                        / Meilisearch
                                        / Typesense
```

#### Search Features

| Feature            | PostgreSQL | Dedicated Engine |
| ------------------ | ---------- | ---------------- |
| Full-text search   | ✅         | ✅               |
| Fuzzy matching     | Limited    | ✅               |
| Autocomplete       | Limited    | ✅               |
| Faceted search     | Manual     | ✅               |
| Typo tolerance     | ❌         | ✅               |
| Multi-language     | ✅         | ✅               |
| Real-time indexing | ✅         | Near real-time   |

#### Search Adapter Pattern

```go
type SearchService interface {
    Index(ctx context.Context, doc Document) error
    Search(ctx context.Context, query SearchQuery) (*SearchResults, error)
    Delete(ctx context.Context, docID string) error
    Suggest(ctx context.Context, prefix string, limit int) ([]Suggestion, error)
}

// Implementations
├── postgres/search.go     # PostgreSQL full-text (default)
├── meilisearch/search.go  # Meilisearch (if needed)
└── memory/search.go       # In-memory (testing)
```

#### What to Index

| Entity     | Searchable Fields                  | Use Case                    |
| ---------- | ---------------------------------- | --------------------------- |
| Menu items | Name, description, category, tags  | Customer/waiter menu search |
| Orders     | Order number, customer name, items | Order lookup                |
| Customers  | Name, email, phone                 | Customer lookup             |
| Invoices   | Invoice number, customer           | Invoice lookup              |

### Infrastructure as Code (IaC)

All infrastructure MUST be defined as code, versioned, and reproducible. The system MUST start
cost-effective on minimal infrastructure and scale progressively as demand grows.

#### IaC Requirements

- All infrastructure MUST be defined in code (no manual cloud console changes)
- Infrastructure changes MUST go through version control and review
- Environments MUST be reproducible from code
- Secrets MUST be managed separately from infrastructure code

#### Progressive Scaling Strategy

```
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 1: Single VM                                                  │
│  ($20-50/month)                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                    Docker Compose                                ││
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   ││
│  │  │ Backend │ │PostgreSQL│ │  Redis  │ │  MinIO  │ │ Caddy/  │   ││
│  │  │   API   │ │         │ │         │ │(storage)│ │ Nginx   │   ││
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼ When needed
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 2: Managed Services                                           │
│  ($100-300/month)                                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │  VM(s)   │  │ Managed  │  │ Managed  │  │    S3    │            │
│  │ Backend  │  │ Postgres │  │  Redis   │  │ Storage  │            │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘            │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼ When needed
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 3: Container Orchestration                                    │
│  ($300+/month)                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │              Kubernetes / Docker Swarm                           ││
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐                            ││
│  │  │ Backend │ │ Backend │ │ Backend │  Auto-scaling              ││
│  │  │   (1)   │ │   (2)   │ │   (N)   │  Load balancing            ││
│  │  └─────────┘ └─────────┘ └─────────┘                            ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

#### Phase 1: Single VM with Docker Compose (Recommended Start)

**Target**: MVP, early customers, 1-10 tenants, <100 concurrent users

**VM Specifications**:
| Provider | Instance | vCPU | RAM | Storage | Est. Cost | Notes |
|----------|----------|------|-----|---------|-----------|-------|
| **AWS Lightsail** | 4GB | 2 | 4GB | 80GB SSD | **$20/mo** | Recommended - AWS ecosystem, easy upgrade |
| AWS Lightsail | 8GB | 2 | 8GB | 160GB SSD | $40/mo | If more RAM needed |
| Hetzner | CX31 | 2 | 8GB | 80GB SSD | €8/mo | Best price, EU data centers |
| DigitalOcean | Basic Droplet | 2 | 4GB | 80GB SSD | $24/mo | Good UX, simple |
| Linode | Linode 4GB | 2 | 4GB | 80GB SSD | $24/mo | Good performance |
| AWS EC2 | t3.medium | 2 | 4GB | 80GB EBS | ~$35/mo | Only if need full AWS features |

**Why AWS Lightsail is Recommended**:

- **Predictable pricing**: Fixed monthly cost, no surprise bills
- **AWS ecosystem access**: Easy to add S3, SES, RDS later without migration
- **Snapshots included**: Automatic backups, easy cloning
- **Upgrade path**: Can migrate to EC2/ECS when needed
- **Managed databases available**: PostgreSQL ($15/mo) and Redis ($10/mo) when ready for Phase 2
- **Load balancers**: $18/mo when horizontal scaling needed
- **CDN integration**: CloudFront for static assets

**Docker Compose Stack**:

```yaml
# infrastructure/docker/docker-compose.prod.yml
version: '3.8'

services:
  backend:
    image: solobueno/backend:${VERSION}
    restart: unless-stopped
    environment:
      - DATABASE_URL=postgres://solobueno:${DB_PASSWORD}@postgres:5432/solobueno
      - REDIS_URL=redis://redis:6379
      - STORAGE_ENDPOINT=http://minio:9000
    depends_on:
      - postgres
      - redis
      - minio
    healthcheck:
      test: ['CMD', 'curl', '-f', 'http://localhost:8080/health']
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    environment:
      - POSTGRES_USER=solobueno
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=solobueno
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U solobueno']
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

  minio:
    image: minio/minio:latest
    restart: unless-stopped
    volumes:
      - minio_data:/data
    environment:
      - MINIO_ROOT_USER=${MINIO_USER}
      - MINIO_ROOT_PASSWORD=${MINIO_PASSWORD}
    command: server /data --console-address ":9001"

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - '80:80'
      - '443:443'
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
  minio_data:
  caddy_data:
  caddy_config:
```

**Caddyfile** (automatic HTTPS):

```
api.solobueno.com {
    reverse_proxy backend:8080
}

app.solobueno.com {
    root * /srv/web
    file_server
    try_files {path} /index.html
}
```

**Deployment Script**:

```bash
#!/bin/bash
# infrastructure/scripts/deploy.sh

set -e

# Pull latest images
docker compose -f docker-compose.prod.yml pull

# Run migrations
docker compose -f docker-compose.prod.yml run --rm backend migrate up

# Deploy with zero-downtime
docker compose -f docker-compose.prod.yml up -d --remove-orphans

# Cleanup old images
docker image prune -f
```

#### Single VM Resource Allocation

```
┌─────────────────────────────────────────────────────────────────┐
│                    4GB RAM VM Allocation                         │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL     │████████████████████│           1.5 GB         │
│  Backend API    │██████████████│                 1.0 GB         │
│  Redis          │████│                           0.5 GB         │
│  MinIO          │████│                           0.5 GB         │
│  Caddy + OS     │████│                           0.5 GB         │
└─────────────────────────────────────────────────────────────────┘
```

#### Backup Strategy (Single VM)

```bash
#!/bin/bash
# infrastructure/scripts/backup.sh
# Run daily via cron: 0 3 * * * /opt/solobueno/backup.sh

DATE=$(date +%Y%m%d)
BACKUP_DIR="/backups"

# PostgreSQL backup
docker compose exec -T postgres pg_dump -U solobueno solobueno | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# MinIO data (optional - if not using external S3)
docker compose exec -T minio mc mirror /data $BACKUP_DIR/minio_$DATE/

# Upload to external storage (optional but recommended)
aws s3 sync $BACKUP_DIR s3://solobueno-backups/ --storage-class STANDARD_IA

# Cleanup old local backups (keep 7 days)
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete
```

#### Phase 2: Managed Services (When Scaling)

**Trigger to move**: Database >50GB, need HA, >50 concurrent users

**AWS Lightsail Managed Services** (recommended upgrade path):
| Component | Self-Hosted (Phase 1) | Lightsail Managed | Est. Cost |
|-----------|----------------------|-------------------|-----------|
| PostgreSQL | Docker | Lightsail Database | $15/mo (1GB) |
| Redis | Docker | Lightsail Database | $10/mo |
| Storage | MinIO | S3 | ~$5/mo |
| Load Balancer | Caddy | Lightsail LB | $18/mo |
| Backend | Single container | Multiple Lightsail instances | $20/mo each |

**Alternative Managed Options**:
| Component | AWS (Full) | DigitalOcean | Other |
|-----------|------------|--------------|-------|
| PostgreSQL | RDS | Managed DB | PlanetScale, Supabase |
| Redis | ElastiCache | Managed Redis | Upstash (serverless) |
| Storage | S3 | Spaces | Cloudflare R2 |

**Benefits**:

- Automatic backups and point-in-time recovery
- High availability (replicas)
- Managed security patches
- Monitoring included
- Stay within AWS ecosystem for easier integration

#### Phase 3: Kubernetes (When Necessary)

**Trigger to move**: Need auto-scaling, >100 concurrent users, multiple services

**Options** (cost-effective to enterprise):
| Option | Cost | Complexity | Best For |
|--------|------|------------|----------|
| Docker Swarm | Low | Low | Simple scaling, small team |
| k3s (lightweight K8s) | Low | Medium | Edge, single-node to small cluster |
| DigitalOcean K8s | Medium | Medium | Managed, predictable cost |
| AWS EKS / GKE | High | High | Enterprise, advanced features |

**K8s Manifests Location**:

```
infrastructure/k8s/
├── base/
│   ├── namespace.yaml
│   ├── backend-deployment.yaml
│   ├── backend-service.yaml
│   ├── configmap.yaml
│   └── secrets.yaml (template)
└── overlays/
    ├── staging/
    │   └── kustomization.yaml
    └── prod/
        ├── kustomization.yaml
        ├── replicas-patch.yaml
        └── resources-patch.yaml
```

#### IaC Tools by Phase (Finalized)

| Phase   | IaC Tool                         | Purpose                                 |
| ------- | -------------------------------- | --------------------------------------- |
| Phase 1 | Docker Compose + Bash scripts    | Container orchestration, deployment     |
| Phase 2 | Terraform (AWS provider)         | Lightsail managed services provisioning |
| Phase 3 | Terraform + Kubernetes manifests | EKS/ECS cluster + app deployment        |

**Note**: Ansible is optional for VM provisioning but Docker Compose handles most needs in Phase 1.

**Terraform Example - AWS Lightsail** (Phase 2):

```hcl
# infrastructure/terraform/lightsail/main.tf

provider "aws" {
  region = "us-east-1"  # Or your preferred region
}

# Lightsail instance for backend
resource "aws_lightsail_instance" "backend" {
  name              = "solobueno-backend"
  availability_zone = "us-east-1a"
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "medium_2_0"  # 2 vCPU, 4GB RAM, $20/mo

  user_data = file("${path.module}/cloud-init.yaml")

  tags = {
    Environment = "production"
  }
}

# Lightsail static IP (so IP doesn't change on restart)
resource "aws_lightsail_static_ip" "backend" {
  name = "solobueno-backend-ip"
}

resource "aws_lightsail_static_ip_attachment" "backend" {
  static_ip_name = aws_lightsail_static_ip.backend.name
  instance_name  = aws_lightsail_instance.backend.name
}

# Lightsail managed PostgreSQL (when ready for Phase 2)
resource "aws_lightsail_database" "postgres" {
  relational_database_name = "solobueno-db"
  availability_zone        = "us-east-1a"
  master_database_name     = "solobueno"
  master_username          = "solobueno"
  master_password          = var.db_password
  blueprint_id             = "postgres_16"
  bundle_id                = "micro_2_0"  # 1 vCPU, 1GB RAM, $15/mo

  skip_final_snapshot = false
  final_snapshot_name = "solobueno-db-final"
}

# S3 bucket for file storage
resource "aws_s3_bucket" "storage" {
  bucket = "solobueno-storage-${var.environment}"
}

resource "aws_s3_bucket_versioning" "storage" {
  bucket = aws_s3_bucket.storage.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

**Lightsail CLI Commands** (alternative to Terraform):

```bash
# Create instance
aws lightsail create-instances \
  --instance-names solobueno-backend \
  --availability-zone us-east-1a \
  --blueprint-id amazon_linux_2 \
  --bundle-id medium_2_0

# Create managed database
aws lightsail create-relational-database \
  --relational-database-name solobueno-db \
  --availability-zone us-east-1a \
  --relational-database-blueprint-id postgres_16 \
  --relational-database-bundle-id micro_2_0 \
  --master-database-name solobueno \
  --master-username solobueno

# Create static IP
aws lightsail allocate-static-ip --static-ip-name solobueno-ip
aws lightsail attach-static-ip --static-ip-name solobueno-ip --instance-name solobueno-backend
```

#### Cost Optimization Guidelines

| Strategy                         | Savings      | Implementation                          |
| -------------------------------- | ------------ | --------------------------------------- |
| Start with single VM             | 80%          | Docker Compose, all-in-one              |
| Use Hetzner/DigitalOcean vs AWS  | 50-70%       | Same Docker images work anywhere        |
| Reserved instances (when stable) | 30-40%       | Commit to 1-year after MVP validated    |
| Managed DB only when needed      | Variable     | Self-host PostgreSQL until scaling pain |
| S3-compatible (MinIO)            | 100% storage | Self-host until egress costs justify S3 |
| Compress backups                 | Storage cost | gzip all backups                        |
| Right-size instances             | Variable     | Monitor and adjust monthly              |

#### Scaling Decision Matrix

| Metric             | Phase 1 OK | Consider Phase 2 | Consider Phase 3 |
| ------------------ | ---------- | ---------------- | ---------------- |
| Tenants            | 1-10       | 10-50            | 50+              |
| Concurrent users   | <100       | 100-500          | 500+             |
| Database size      | <20GB      | 20-100GB         | 100GB+           |
| Requests/sec       | <50        | 50-200           | 200+             |
| Uptime requirement | 99%        | 99.5%            | 99.9%            |
| Team size          | 1-2        | 2-5              | 5+               |

#### Monitoring (All Phases)

Even on single VM, basic monitoring is essential:

**Phase 1 (Free/Low-Cost)**:

- Uptime: UptimeRobot (free), Healthchecks.io (free)
- Logs: Docker logs + Loki (self-hosted) or Papertrail (free tier)
- Metrics: Prometheus + Grafana (self-hosted, ~200MB RAM)

**Phase 2+**:

- Full observability stack or managed (Datadog, New Relic)

```yaml
# Add to docker-compose for basic monitoring
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  profiles: ['monitoring'] # Optional, enable with --profile monitoring

grafana:
  image: grafana/grafana:latest
  volumes:
    - grafana_data:/var/lib/grafana
  profiles: ['monitoring']
```

**Rationale**: Budget constraints require starting lean. A single well-configured VM can handle
significant load for a restaurant ERP. Docker Compose provides container benefits without
Kubernetes complexity. Infrastructure as code ensures reproducibility and enables smooth
scaling when business growth justifies increased infrastructure investment.

### Observability Infrastructure

```
backend/internal/shared/observability/
├── logger.go               # Logger interface
├── zerolog/                # zerolog implementation
│   └── logger.go
├── metrics.go              # Metrics interface (no-op initially)
├── tracer.go               # Tracer interface (no-op initially)
└── middleware.go           # HTTP/GraphQL logging middleware
```

### Event Infrastructure

```
backend/internal/shared/events/
├── event.go                # Base event interface and types
├── bus.go                  # EventBus interface
├── memory/                 # In-memory implementation
│   └── bus.go
├── nats/                   # NATS implementation (prod)
│   └── bus.go
└── store/                  # Event store for replay/audit (optional)
    └── postgres.go
```

### Platform Targets

- **Mobile App**: iOS 14+ and Android 10+ (React Native)
- **Kitchen Display**: Tablet (React Native or web)
- **Backoffice**: Desktop browsers (Chrome, Firefox, Safari)
- **Admin Portal**: Desktop browsers with responsive design

### Performance Constraints

- Mobile app MUST remain responsive with 500+ menu items loaded
- Order creation MUST complete in <500ms (local) or <2s (synced)
- GraphQL queries MUST resolve in <100ms (p95) for typical operations
- Event propagation MUST complete in <50ms (in-memory) or <200ms (message broker)
- Backoffice reports MUST render in <5 seconds for 30-day data ranges
- Offline storage MUST support 7 days of operations without sync
- Analytics event upload MUST NOT block UI operations

### Security Requirements

- All API endpoints MUST require authentication (JWT with tenant context)
- Role-based access control (RBAC) per tenant (waiter/kitchen/manager/admin/owner)
- Payment card data MUST NOT be stored (use payment processor tokens)
- Tenant isolation MUST be enforced at module/repository layer
- Events MUST NOT contain sensitive PII or payment data
- Logs MUST NOT contain sensitive data (passwords, tokens, full card numbers)
- Audit logs MUST track all financial transactions per tenant

## Database Migrations

### Migration Tooling

The project MUST use **golang-migrate** or **goose** for database migrations:

| Tool               | Pros                                       | Recommendation                               |
| ------------------ | ------------------------------------------ | -------------------------------------------- |
| **golang-migrate** | Library + CLI, driver support, widely used | Default choice                               |
| **goose**          | Embedded Go migrations, simpler            | Alternative if Go-based migrations preferred |

### Migration File Structure

```
backend/migrations/
├── 000001_create_tenants.up.sql
├── 000001_create_tenants.down.sql
├── 000002_create_users.up.sql
├── 000002_create_users.down.sql
├── ...
├── 0000XX_create_feedback.up.sql
├── 0000XX_create_feedback.down.sql
├── 0000XX_create_analytics_events.up.sql
├── 0000XX_create_analytics_events.down.sql
└── ...
```

**Naming Convention**: `{version}_{description}.{direction}.sql`

### Migration CLI Commands

```bash
make migrate-up ENV=dev         # Apply all pending
make migrate-down ENV=dev STEPS=1   # Revert last
make migrate-version ENV=dev    # Check version
make migrate-create NAME=...    # Create new pair
make migrate-validate           # CI validation
```

### Migration Rules

- Every migration MUST have a revert (`.down.sql`)
- Migrations MUST be backward-compatible
- Zero-downtime pattern for breaking changes
- Large data changes MUST be batched

## Environments & Deployment

### Environment Definitions

| Environment | Purpose                          | Data                      | Access              |
| ----------- | -------------------------------- | ------------------------- | ------------------- |
| **dev**     | Local development                | Seed/mock data            | Developers only     |
| **test**    | Automated testing, QA validation | Synthetic test data       | Team + CI           |
| **staging** | Pre-production validation, UAT   | Anonymized prod-like data | Team + stakeholders |
| **prod**    | Live production                  | Real customer data        | End users           |

### Deployment Pipeline

```
Push → Build → Test → Deploy Test → Deploy Staging → [Approval] → Deploy Prod
                           │              │                            │
                           ▼              ▼                            ▼
                      Migrations     Migrations                   Migrations
```

### Promotion Rules

| Promotion      | Trigger         | Gates                                         |
| -------------- | --------------- | --------------------------------------------- |
| main → test    | Automatic       | Build passes, tests pass, migrations pass     |
| test → staging | Automatic       | E2E tests pass, no critical bugs              |
| staging → prod | Manual approval | UAT sign-off, security review (if applicable) |

## Development Workflow

### Code Review Requirements

- All changes MUST be reviewed by at least one team member before merge
- PRs affecting payment, fiscal plugins, or tenant isolation MUST have two reviewers
- New domain events MUST be reviewed for schema compatibility
- Migration files MUST be reviewed for revert safety and backward compatibility
- No direct commits to `main` branch

### Testing Gates

- CI MUST pass all unit and integration tests before merge
- Module contract tests MUST pass for any interface changes
- Migration tests MUST pass (apply, revert, re-apply cycle)
- Event handler idempotency tests MUST pass
- Plugin compliance tests MUST pass for any plugin changes
- E2E tests MUST pass in test environment before staging promotion
- Manual QA required for UI changes affecting mobile workflows

### Documentation Standards

- GraphQL schema MUST be self-documenting with descriptions
- REST endpoints MUST have OpenAPI documentation
- Module interfaces MUST have Go doc comments
- Domain events MUST be documented in `docs/events/` catalog
- Analytics events MUST be documented with purpose and key fields
- Plugin contracts MUST have implementation guides
- Architecture decisions MUST be recorded in `docs/adr/`
- Runbooks MUST exist for common operational tasks

## Governance

This constitution represents the non-negotiable principles governing Solobueno ERP development.

### Amendment Process

1. Proposed changes MUST be documented with rationale
2. Changes to Core Principles require team consensus
3. All amendments MUST update the version number and Last Amended date
4. Dependent templates MUST be reviewed for consistency after amendments

### Versioning Policy

- **MAJOR**: Principle removal, redefinition, or backward-incompatible governance change
- **MINOR**: New principle added, section materially expanded
- **PATCH**: Clarifications, wording improvements, typo fixes

### Compliance Expectations

- All PRs MUST verify alignment with Core Principles
- Plan documents MUST include a Constitution Check section
- Complexity beyond these principles MUST be justified in writing
- Plugin implementations MUST pass contract compliance tests
- New modules MUST follow the established module design pattern
- New events MUST follow event design rules and be added to the event catalog
- New migrations MUST have working up AND down scripts
- Logs MUST follow structured logging standards
- Deployments MUST follow the defined pipeline stages

## Appendix: Finalized Technology Decisions

All technology choices have been finalized to eliminate ambiguity and enable consistent development.

### Core Stack

| Category             | Decision                  | Alternative Considered |
| -------------------- | ------------------------- | ---------------------- |
| **Cloud Provider**   | AWS Lightsail             | Hetzner, DigitalOcean  |
| **Region**           | us-east-1                 | -                      |
| **Backend Language** | Go 1.22+                  | -                      |
| **Frontend**         | React Native + TypeScript | -                      |
| **Database**         | PostgreSQL 16             | -                      |
| **Cache**            | Redis 7                   | -                      |

### Backend Libraries

| Category           | Decision       | Alternative Considered |
| ------------------ | -------------- | ---------------------- |
| **GraphQL**        | gqlgen         | -                      |
| **REST Framework** | Chi            | Echo                   |
| **Migrations**     | golang-migrate | goose                  |
| **Logging**        | zerolog        | zap                    |
| **Job Queue**      | Asynq          | custom                 |

### Infrastructure

| Category               | Decision                | Alternative Considered    |
| ---------------------- | ----------------------- | ------------------------- |
| **Event Bus (Prod)**   | NATS                    | Redis Streams, Kafka      |
| **File Storage**       | AWS S3                  | MinIO, GCS                |
| **Email**              | AWS SES                 | SendGrid                  |
| **Secrets**            | AWS SSM Parameter Store | Vault, GCP Secret Manager |
| **Container Registry** | Amazon ECR              | GitHub, GitLab            |

### Frontend / Mobile

| Category          | Decision       | Alternative Considered |
| ----------------- | -------------- | ---------------------- |
| **Offline Sync**  | WatermelonDB   | custom                 |
| **i18n Strategy** | Package-driven | Database-driven        |
| **Monorepo Tool** | Turborepo      | Nx                     |

### DevOps

| Category                  | Decision       | Alternative Considered        |
| ------------------------- | -------------- | ----------------------------- |
| **CI/CD**                 | GitHub Actions | GitLab CI                     |
| **Phase 1 Orchestration** | Docker Compose | -                             |
| **Phase 2+ IaC**          | Terraform      | Pulumi                        |
| **Search (Phase 1)**      | PostgreSQL FTS | Meilisearch (later if needed) |

### Languages Supported

| Language                | Code   | Status    |
| ----------------------- | ------ | --------- |
| Spanish (Latin America) | es-419 | Primary   |
| English                 | en     | Secondary |

**Version**: 1.7.0 | **Ratified**: 2025-01-29 | **Last Amended**: 2025-01-29
