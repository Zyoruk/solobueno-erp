## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:

- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost). The `.husky/pre-commit` hook also runs this automatically before every commit (no-ops if `graphify` isn't installed), so a stale graph should only ever be a mid-session thing, never a committed one.

## Dev cycle (backend/Go)

Before calling any backend code change done:

- `cd backend && go build ./... && go vet ./...`
- `go test ./...` (or scope to the touched package)
- Every backend module (`internal/<module>/`) with business logic MUST have an `e2e_test.go` driving its real router/handlers over HTTP (`httptest`), not just unit tests. If the touched module doesn't have one yet, add it.
- Run e2e tests with `make backend-test-e2e` (runs every module's `TestE2E*`; `MODULE=<name>` scopes to one, `COUNT=N` repeats)
- New non-trivial logic (branch, service method, handler) gets a unit test in the same change, not deferred
- Every new/changed REST handler MUST get swaggo/swag annotations (`@Summary`/`@Router`/etc. — see `backend/internal/auth/handler/*.go` for examples) and `make backend-swagger` regenerated in the same change, per Constitution III's OpenAPI 3.0 requirement
