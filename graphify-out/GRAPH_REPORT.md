# Graph Report - . (2026-08-06)

## Corpus Check

- 131 files · ~74,264 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary

- 840 nodes · 983 edges · 108 communities (51 shown, 57 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 45 edges (avg confidence: 0.82)
- Token cost: 460,312 input · 0 output

## Community Hubs (Navigation)

- CI Pipeline & Coverage Gates
- Speckit Planning Workflow
- Caveman-Compress CLI
- Turborepo Task Config
- Caveman Commit & Compress Skills
- Caveman-Compress Benchmark Tool
- Base TypeScript Config
- Root Lint & Format Tooling
- Backoffice App Package
- Mobile App Package
- Docker Local Dev Services
- Root Workspace Package Config
- Admin App Package
- Kitchen-Display App Package
- GraphQL Client Package
- Speckit Agent-Context Script
- Core Domain Entities (Auth/Menu/Orders)
- Analytics Package Config
- i18n Package Config
- Types Package Config
- UI Package Config
- Menu Domain Model
- Speckit Prerequisite Checks
- Analytics Tracking Client
- Admin App TSConfig
- Backoffice App TSConfig
- Kitchen-Display App TSConfig
- Mobile App TSConfig
- UI Package TSConfig
- Backend Domain Modules Overview
- Analytics Package TSConfig
- GraphQL Client TSConfig
- i18n Package TSConfig
- Types Package TSConfig
- Cavecrew Subagent Family
- Go Version Package
- Speckit Branch Creation Script
- Speckit Constitution & Analyze Flow
- i18n Locale Runtime
- Speckit Checklist Gate
- GraphQL Client Entry (Stub)
- Admin App Entry (Stub)
- Backoffice App Entry (Stub)
- Kitchen-Display App Entry (Stub)
- Mobile App Entry (Stub)
- Speckit Spec & Checklist Templates
- Speckit Clarify Handoff
- Speckit Tasks & Issues Commands
- UI Package Entry (Stub)
- Observability & Feedback Principles
- Caveman-Compress Package Init
- Docker Health-Check Script
- Notification & Report Templates
- Vitest Coverage Config
- Go Module Root
- User-Centric UI Principle
- Auth Role Entity
- Auth Session Entity
- Auth Tenant Entity
- Auth User Entity
- Config Branding Settings
- Config Business Settings
- Config Feature Flag
- Config Global Settings
- Config Tenant Settings
- Analytics Tracker Concept
- GraphQL Client Concept
- i18n Locale Files Concept
- Shared Packages Concept
- i18n Translation Function
- Shared Type Definitions
- Shared UI Components
- Menu Category Entity
- Menu Item Image Entity
- Menu Modifier Entity
- Menu Modifier Group Entity
- Tables Reservation Entity
- Tables Section Entity
- Tables Server Assignment
- Tables Session Entity
- Orders Check Entity
- Orders Kitchen Ticket
- Orders Modification Entity
- Inventory Ingredient Category
- Inventory Supplier Entity
- Payments Cash Drawer
- Payments Method Entity
- Payments Plugin Interface
- Payments Terminal Entity
- Payments Refund Entity
- Billing Plugin Interface
- Billing Credit Note
- Billing Invoice Line
- Billing Tax Rate
- Notifications Delivery Attempt
- Notifications Channel Entity
- Notifications Preference Entity
- Analytics Dashboard Entity
- Analytics Metric Definition
- Analytics Report Entity
- Feedback Complaint Category
- Feedback Response Entity
- Reporting Export Entity
- Reporting Permission Entity
- Reporting Scheduled Report

## God Nodes (most connected - your core abstractions)

1. `compilerOptions` - 19 edges
2. `compress_file()` - 15 edges
3. `validate()` - 14 edges
4. `ci.yml GitHub Actions Workflow` - 12 edges
5. `scripts` - 11 edges
6. `main()` - 10 edges
7. `Tasks: Initialize Monorepo Structure` - 10 edges
8. `Feature Specification: Docker Local Development Environment` - 10 edges
9. `detect_file_type()` - 9 edges
10. `log_info()` - 9 edges

## Surprising Connections (you probably didn't know these)

- `Graphify Project Instructions (CLAUDE.md)` --semantically_similar_to--> `Agent Context File Template` [INFERRED] [semantically similar]
  CLAUDE.md → .specify/templates/agent-file-template.md
- `CI GitHub Actions Workflow` --implements--> `Test-Driven Development` [INFERRED]
  .github/workflows/ci.yml → .specify/memory/constitution.md
- `Solobueno ERP README` --references--> `Mobile-First for Operations` [INFERRED]
  README.md → .specify/memory/constitution.md
- `Solobueno ERP README` --references--> `Offline-First Operations` [INFERRED]
  README.md → .specify/memory/constitution.md
- `Solobueno ERP README` --references--> `Plugin-Driven Architecture` [INFERRED]
  README.md → .specify/memory/constitution.md

## Import Cycles

- None detected.

## Hyperedges (group relationships)

- **Caveman toolkit skill family** — \_agents_skills_caveman_readme_caveman, \_agents_skills_caveman_commit_readme_caveman_commit, \_agents_skills_caveman_compress_readme_caveman_compress, \_agents_skills_caveman_help_readme_caveman_help, \_agents_skills_caveman_review_readme_caveman_review, \_agents_skills_caveman_stats_readme_caveman_stats, \_agents_skills_cavecrew_readme_cavecrew [INFERRED 0.85]
- **cavecrew locate-fix-verify delegation chain** — \_agents_skills_cavecrew_readme_cavecrew_investigator, \_agents_skills_cavecrew_readme_cavecrew_builder, \_agents_skills_cavecrew_readme_cavecrew_reviewer [EXTRACTED 1.00]
- **speckit spec-driven development pipeline** — \_cursor_commands_speckit_constitution_speckit_constitution, \_cursor_commands_speckit_clarify_speckit_clarify, \_cursor_commands_speckit_checklist_speckit_checklist, \_cursor_commands_speckit_analyze_speckit_analyze, \_cursor_commands_speckit_implement_speckit_implement [INFERRED 0.85]
- **Frontline Operations UX Pattern** — specify_memory_constitution_mobilefirstforoperations, specify_memory_constitution_offlinefirstoperations, specify_memory_constitution_usercentricinterfacedesign [INFERRED 0.85]
- **Multi-Country Expansion Strategy** — specify_memory_constitution_plugindrivenarchitecture, specify_memory_constitution_i18npluggablecompliance, specify_memory_constitution_whitelabelmultitenant [INFERRED 0.85]
- **001-init-monorepo Feature Artifact Bundle** — specs_001_init_monorepo_checklists_requirements_speckitchecklist, specs_001_init_monorepo_contracts_readme_contractsoverview, specs_001_init_monorepo_data_model_datamodeldoc [EXTRACTED 1.00]
- **CI Pipeline Job Flow (lint to build to test)** — ci_workflow_ci_yml, lint_job, build_job, test_job [EXTRACTED 1.00]
- **Local Development Services Stack** — postgresql_container, redis_container, minio_container, docker_network_solobueno, docker_volumes_named [EXTRACTED 1.00]
- **Security Scanning and Merge Gating** — codeql_workflow_codeql_yml, dependabot_config, specs_003_ci_pipeline_branch_protection_rules_main [EXTRACTED 1.00]
- **Pluggable Integration Pattern (Payments, Billing, Notifications)** — specs_011_payments_module_spec_paymentplugin, specs_012_billing_module_spec_billingplugin, specs_013_notifications_module_spec_notificationchannel [INFERRED 0.75]
- **Order as Cross-Module Reference Hub** — specs_009_orders_module_spec_order, specs_011_payments_module_spec_payment, specs_012_billing_module_spec_invoice, specs_015_feedback_module_spec_rating [INFERRED 0.85]
- **Reporting Module Aggregates All Business Modules** — specs_016_reporting_module_spec_reporting_module, specs_009_orders_module_spec_orders_module, specs_010_inventory_module_spec_inventory_module, specs_011_payments_module_spec_payments_module, specs_012_billing_module_spec_billing_module, specs_014_analytics_module_spec_analytics_module, specs_015_feedback_module_spec_feedback_module [EXTRACTED 1.00]

## Communities (108 total, 57 thin omitted)

### Community 0 - "CI Pipeline & Coverage Gates"

Cohesion: 0.12
Nodes (40): Build Job (ci.yml), Multi-layer CI Caching Strategy, ci.yml GitHub Actions Workflow, GitHub Actions Initial CI Setup, 80% Code Coverage Threshold, Prettier + ESLint + golangci-lint, CODEOWNERS File, CodeQL Analyze Job (+32 more)

### Community 1 - "Speckit Planning Workflow"

Cohesion: 0.07
Nodes (40): Graphify Project Instructions (CLAUDE.md), Phase 0: Outline & Research, Phase 1: Design & Contracts, speckit.plan Command, Dependabot Configuration, CI GitHub Actions Workflow, Go Coverage Threshold (interim 10%), CodeQL Security Analysis Workflow (+32 more)

### Community 2 - "Caveman-Compress CLI"

Cohesion: 0.10
Nodes (33): main(), print_usage(), backup_dir_for(), build_compress_prompt(), build_fix_prompt(), call_claude(), compress_file(), first_nonblank_line() (+25 more)

### Community 3 - "Turborepo Task Config"

Cohesion: 0.07
Nodes (31): .env, **/.env.\*local, .next/**, NODE_ENV, \*.tsbuildinfo, dependsOn, env, outputs (+23 more)

### Community 4 - "Caveman Commit & Compress Skills"

Cohesion: 0.08
Nodes (29): caveman-commit (skill overview), Conventional Commits format, Auto-Clarity rule (commit), caveman-commit (SKILL instructions), out-of-tree backup dir (README), caveman-compress (skill overview), caveman-compress (SECURITY doc), Snyk High Risk rating (false positive) (+21 more)

### Community 5 - "Caveman-Compress Benchmark Tool"

Cohesion: 0.15
Nodes (23): benchmark_pair(), count_tokens(), main(), print_table(), Path, count_bullets(), extract_code_blocks(), extract_headings() (+15 more)

### Community 6 - "Base TypeScript Config"

Cohesion: 0.08
Nodes (25): ES2022, compilerOptions, declaration, declarationMap, esModuleInterop, forceConsistentCasingInFileNames, isolatedModules, lib (+17 more)

### Community 7 - "Root Lint & Format Tooling"

Cohesion: 0.08
Nodes (25): eslint, eslint-config-prettier, husky, lint-staged, devDependencies, eslint, eslint-config-prettier, husky (+17 more)

### Community 8 - "Backoffice App Package"

Cohesion: 0.08
Nodes (23): dependencies, @solobueno/analytics, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript (+15 more)

### Community 9 - "Mobile App Package"

Cohesion: 0.08
Nodes (23): dependencies, @solobueno/analytics, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript (+15 more)

### Community 10 - "Docker Local Dev Services"

Cohesion: 0.17
Nodes (24): Named Volume Data Persistence Strategy, Simple Development Credentials, Enhanced docker-compose.yml (002), solobueno-network Docker Network, Named Docker Volumes (postgres/redis/minio data), health-check.sh Script, Native Docker Health Check Strategy, Makefile docker-\* Targets (+16 more)

### Community 11 - "Root Workspace Package Config"

Cohesion: 0.09
Nodes (22): engines, node, pnpm, lint-staged, _.{json,md,yml,yaml}, _.{ts,tsx,js,jsx}, name, packageManager (+14 more)

### Community 12 - "Admin App Package"

Cohesion: 0.09
Nodes (21): dependencies, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript, @solobueno/graphql (+13 more)

### Community 13 - "Kitchen-Display App Package"

Cohesion: 0.09
Nodes (21): dependencies, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript, @solobueno/graphql (+13 more)

### Community 14 - "GraphQL Client Package"

Cohesion: 0.09
Nodes (21): dependencies, @solobueno/types, devDependencies, tsup, typescript, exports, @solobueno/types, tsup (+13 more)

### Community 15 - "Speckit Agent-Context Script"

Cohesion: 0.23
Nodes (14): create_new_agent_file(), log_error(), log_info(), log_success(), log_warning(), main(), parse_plan_data(), print_summary() (+6 more)

### Community 16 - "Core Domain Entities (Auth/Menu/Orders)"

Cohesion: 0.12
Nodes (20): AuthEvent, ConfigChange, MenuItem, Table, Order, OrderItem, Ingredient, Recipe (+12 more)

### Community 17 - "Analytics Package Config"

Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 18 - "i18n Package Config"

Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 19 - "Types Package Config"

Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 20 - "UI Package Config"

Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 21 - "Menu Domain Model"

Cohesion: 0.25
Nodes (14): BaseEntity, MenuCategory, MenuItem, Order, OrderItem, OrderStatus, Payment, PaymentMethod (+6 more)

### Community 22 - "Speckit Prerequisite Checks"

Cohesion: 0.18
Nodes (8): check-prerequisites.sh script, check_dir(), check_feature_branch(), check_file(), get_feature_paths(), has_git(), common.sh script, setup-plan.sh script

### Community 23 - "Analytics Tracking Client"

Cohesion: 0.27
Nodes (11): AnalyticsEvent, eventQueue, generateId(), getQueuedEvents(), initAnalytics(), QueuedEvent, TODO: Process queued events, TODO: Send to analytics service (+3 more)

### Community 24 - "Admin App TSConfig"

Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 25 - "Backoffice App TSConfig"

Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 26 - "Kitchen-Display App TSConfig"

Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 27 - "Mobile App TSConfig"

Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 28 - "UI Package TSConfig"

Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 29 - "Backend Domain Modules Overview"

Cohesion: 0.30
Nodes (12): Authentication Module, Configuration Module, Menu Module, Tables Module, Orders Module, Inventory Module, Payments Module, Billing Module (+4 more)

### Community 30 - "Analytics Package TSConfig"

Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 31 - "GraphQL Client TSConfig"

Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 32 - "i18n Package TSConfig"

Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 33 - "Types Package TSConfig"

Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 34 - "Cavecrew Subagent Family"

Cohesion: 0.31
Nodes (9): cavecrew (skill overview), cavecrew-builder subagent, cavecrew-investigator subagent, cavecrew-reviewer subagent, Caveman toolkit root README, cavecrew (SKILL instructions), Code Reviewer (vanilla agent), Explore (vanilla agent) (+1 more)

### Community 35 - "Go Version Package"

Cohesion: 0.33
Nodes (7): Info(), IsPreRelease(), TestAppName(), TestInfo(), TestIsPreRelease(), TestVersion(), T

### Community 37 - "Speckit Constitution & Analyze Flow"

Cohesion: 0.29
Nodes (7): .specify/scripts/bash/check-prerequisites.sh, .specify/memory/constitution.md, speckit.analyze command, .specify/memory/constitution.md (target), speckit.constitution command, handoff to speckit.specify, Sync Impact Report

### Community 38 - "i18n Locale Runtime"

Cohesion: 0.43
Nodes (5): getLocale(), Locale, locales, setLocale(), t()

### Community 39 - "Speckit Checklist Gate"

Cohesion: 0.33
Nodes (6): .specify/templates/checklist-template.md, speckit.checklist command, "Unit Tests for English" principle, FEATURE_DIR/checklists/ gate, speckit.implement command, speckit.tasks command (referenced)

### Community 40 - "GraphQL Client Entry (Stub)"

Cohesion: 0.33
Nodes (4): GRAPHQL_VERSION, GraphQLClientConfig, TODO: Generated GraphQL operations will be exported here, TODO: Implement GraphQL client

### Community 41 - "Admin App Entry (Stub)"

Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React admin portal

### Community 42 - "Backoffice App Entry (Stub)"

Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React web app

### Community 43 - "Kitchen-Display App Entry (Stub)"

Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React Native kitchen display

### Community 44 - "Mobile App Entry (Stub)"

Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React Native app entry point

### Community 45 - "Speckit Spec & Checklist Templates"

Cohesion: 0.67
Nodes (4): speckit.specify Command, Checklist Template, Feature Specification Template, Init Monorepo Spec Quality Checklist

### Community 46 - "Speckit Clarify Handoff"

Cohesion: 0.67
Nodes (3): ## Clarifications spec section, speckit.clarify command, handoff to speckit.plan

### Community 47 - "Speckit Tasks & Issues Commands"

Cohesion: 0.67
Nodes (3): speckit.tasks Command, speckit.taskstoissues Command, Tasks Template

### Community 49 - "Observability & Feedback Principles"

Cohesion: 0.67
Nodes (3): Clickstream & Analytics Events, Feedback Module, Observability & Monitoring

## Ambiguous Edges - Review These

- `Prettier + ESLint + golangci-lint` → `Lint Job (ci.yml)` [AMBIGUOUS]
  specs/003-ci-pipeline/data-model.md · relation: implements

## Knowledge Gaps

- **362 isolated node(s):** `common.sh script`, `create-new-feature.sh script`, `SPECIFY_FEATURE`, `name`, `version` (+357 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **57 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions

_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Prettier + ESLint + golangci-lint` and `Lint Job (ci.yml)`?**
  _Edge tagged AMBIGUOUS (relation: implements) - confidence is low._
- **Why does `Feature Specification: Initialize Monorepo Structure` connect `CI Pipeline & Coverage Gates` to `Docker Local Dev Services`?**
  _High betweenness centrality (0.003) - this node is a cross-community bridge._
- **Why does `Feature Specification: Docker Local Development Environment` connect `Docker Local Dev Services` to `CI Pipeline & Coverage Gates`?**
  _High betweenness centrality (0.002) - this node is a cross-community bridge._
- **Why does `devDependencies` connect `Root Lint & Format Tooling` to `Root Workspace Package Config`?**
  _High betweenness centrality (0.002) - this node is a cross-community bridge._
- **What connects `common.sh script`, `create-new-feature.sh script`, `SPECIFY_FEATURE` to the rest of the system?**
  _362 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `CI Pipeline & Coverage Gates` be split into smaller, more focused modules?**
  _Cohesion score 0.11538461538461539 - nodes in this community are weakly interconnected._
- **Should `Speckit Planning Workflow` be split into smaller, more focused modules?**
  _Cohesion score 0.06538461538461539 - nodes in this community are weakly interconnected._
