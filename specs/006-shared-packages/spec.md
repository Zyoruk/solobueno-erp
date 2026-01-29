# Feature Specification: Shared Packages

**Feature Branch**: `006-shared-packages`  
**Created**: 2025-01-29  
**Status**: Draft  
**Dependencies**: 001-init-monorepo

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Developer Uses Shared UI Components (Priority: P1)

As a frontend developer, I want to use pre-built UI components from a shared library, so that I can build consistent interfaces quickly without duplicating code.

**Why this priority**: Shared components ensure consistency and accelerate development across all apps.

**Independent Test**: Can be fully tested by importing and rendering components in any app.

**Acceptance Scenarios**:

1. **Given** a developer imports a Button component, **When** they render it in the mobile app, **Then** it displays with correct styling and behavior.

2. **Given** the same Button component is used in the web app, **When** rendered, **Then** it has consistent appearance matching the mobile version.

3. **Given** a developer needs a form input, **When** they import TextInput from the shared library, **Then** it includes validation styling and accessibility features.

---

### User Story 2 - Developer Uses Translated Strings (Priority: P1)

As a developer building user interfaces, I want to access translated strings from a shared package, so that the app supports multiple languages consistently.

**Why this priority**: All user-facing text must be translatable; this is foundational for i18n.

**Independent Test**: Can be fully tested by switching locale and verifying strings change.

**Acceptance Scenarios**:

1. **Given** a developer uses the translation function, **When** they pass a key like "common.save", **Then** they receive the translated string for the current locale.

2. **Given** the app locale is Spanish, **When** translated strings are displayed, **Then** all text appears in Spanish.

3. **Given** a translation key is missing, **When** the translation function is called, **Then** it returns the key itself and logs a warning.

---

### User Story 3 - Developer Uses Shared TypeScript Types (Priority: P1)

As a developer working across frontend and backend, I want shared type definitions, so that data structures are consistent and type-safe throughout the system.

**Why this priority**: Type consistency prevents bugs and improves developer experience.

**Independent Test**: Can be fully tested by importing types and using them in both frontend and backend.

**Acceptance Scenarios**:

1. **Given** an Order type is defined in shared types, **When** a developer imports it in the mobile app, **Then** TypeScript enforces the correct structure.

2. **Given** the Order type changes, **When** the shared package is rebuilt, **Then** TypeScript errors appear in any code using the old structure.

3. **Given** a developer creates a new entity, **When** they add the type to the shared package, **Then** it's available to all apps after rebuild.

---

### User Story 4 - Developer Uses Generated GraphQL Client (Priority: P2)

As a frontend developer, I want a type-safe GraphQL client generated from the schema, so that I can make API calls without manual type definitions.

**Why this priority**: Generated clients eliminate type mismatches between frontend and backend.

**Independent Test**: Can be fully tested by making GraphQL queries and verifying type safety.

**Acceptance Scenarios**:

1. **Given** a GraphQL query is defined, **When** the client is generated, **Then** the query function has correct TypeScript types for variables and response.

2. **Given** the GraphQL schema changes, **When** codegen runs, **Then** the client is updated and type errors show if usage is incompatible.

3. **Given** a developer writes an invalid query, **When** TypeScript checks the code, **Then** errors are reported at compile time.

---

### User Story 5 - Developer Uses Analytics Helpers (Priority: P3)

As a frontend developer, I want helper functions for tracking analytics events, so that I can instrument the app consistently without boilerplate.

**Why this priority**: Analytics helpers standardize event tracking but are not blocking for core features.

**Independent Test**: Can be fully tested by calling track functions and verifying events are queued.

**Acceptance Scenarios**:

1. **Given** a developer calls trackEvent with event data, **When** the call completes, **Then** the event is queued for upload.

2. **Given** the app is offline, **When** track functions are called, **Then** events are stored locally for later upload.

3. **Given** an event has required fields, **When** a developer omits one, **Then** TypeScript reports a compile-time error.

---

### Edge Cases

- What happens when a package has a build error? Other packages should still build if they don't depend on it.
- What happens when translations are incomplete? Missing keys return the key itself; complete languages are documented.
- What happens when packages have version mismatches? pnpm workspace ensures all packages use the same version.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST provide @solobueno/ui package with shared React/React Native components.

- **FR-002**: UI package MUST include these core components: Button, TextInput, Card, Modal, List, Icon.

- **FR-003**: UI components MUST be compatible with both React Native and React web.

- **FR-004**: System MUST provide @solobueno/i18n package with translation utilities.

- **FR-005**: i18n package MUST support Spanish (es-419) and English (en) locales.

- **FR-006**: i18n package MUST provide a type-safe translation function with key autocompletion.

- **FR-007**: System MUST provide @solobueno/types package with shared TypeScript definitions.

- **FR-008**: Types package MUST include definitions for all domain entities (Order, User, MenuItem, etc.).

- **FR-009**: System MUST provide @solobueno/graphql package with generated GraphQL client.

- **FR-010**: GraphQL package MUST be generated from the backend schema automatically.

- **FR-011**: System MUST provide @solobueno/analytics package with event tracking utilities.

- **FR-012**: Analytics package MUST support offline event queuing.

- **FR-013**: All packages MUST be buildable independently with Turborepo.

- **FR-014**: All packages MUST have TypeScript strict mode enabled.

### Key Entities

- **UI Components**: Reusable React/React Native components for consistent interfaces.
- **Translation Function**: Utility that returns localized strings based on current locale.
- **Locale Files**: JSON files containing translated strings for each supported language.
- **Type Definitions**: TypeScript interfaces and types shared across applications.
- **GraphQL Client**: Generated code for making type-safe API calls.
- **Analytics Tracker**: Utility for recording and queuing analytics events.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: All UI components render correctly on iOS, Android, and web browsers.

- **SC-002**: Translation coverage reaches 100% for Spanish and English on all user-facing strings.

- **SC-003**: TypeScript compilation catches 100% of type mismatches when using shared types.

- **SC-004**: GraphQL client generation completes within 10 seconds after schema changes.

- **SC-005**: Packages build independently in under 30 seconds each.

- **SC-006**: Import statements use package names (@solobueno/\*) not relative paths.

- **SC-007**: All packages pass linting with zero warnings.
