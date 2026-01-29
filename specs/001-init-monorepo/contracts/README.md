# Contracts: Initialize Monorepo Structure

**Feature**: 001-init-monorepo  
**Date**: 2025-01-29

## Overview

This feature establishes the project structure and does not introduce API contracts. API contracts (GraphQL schemas, REST endpoints) will be defined in subsequent features.

## Conventions Established

### Package Naming Convention

All shared packages use the `@solobueno/` scope:

| Package              | Name                   | Purpose                              |
| -------------------- | ---------------------- | ------------------------------------ |
| UI Components        | `@solobueno/ui`        | Shared React/React Native components |
| Internationalization | `@solobueno/i18n`      | Translation strings and utilities    |
| Type Definitions     | `@solobueno/types`     | Shared TypeScript types              |
| GraphQL Client       | `@solobueno/graphql`   | Generated GraphQL client             |
| Analytics            | `@solobueno/analytics` | Client-side analytics helpers        |

### Import Conventions

Packages should be imported using their scoped names:

```typescript
// Good
import { Button } from '@solobueno/ui';
import { t } from '@solobueno/i18n';
import type { Order } from '@solobueno/types';

// Avoid relative paths across packages
import { Button } from '../../../packages/ui'; // Bad
```

### Go Module Import Convention

Internal packages use the module path:

```go
// Good
import "github.com/solobueno/erp/internal/orders"
import "github.com/solobueno/erp/internal/shared/types"

// Internal packages cannot be imported from outside
```

## Future Contracts

The following contracts will be defined in subsequent features:

- **GraphQL Schema**: `backend/api/graphql/schema.graphql`
- **REST OpenAPI**: `backend/api/rest/openapi.yaml`
- **Event Schemas**: `docs/events/*.md`
