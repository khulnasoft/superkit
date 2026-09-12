# Superkit Project Plan

## Executive summary

Superkit already has a strong foundation for a lightweight Go web framework: modular app structure, server-rendered views, auth flows, validation utilities, eventing, and a bootstrap example app. The biggest gap is not raw feature count but production maturity, security design, and framework ergonomics.

Current estimate: 3.5/5
Target state: 5/5

## Product objective

Build a framework that is fast to start with, secure by default, easy to extend, and mature enough for real production use.

## Strategic priorities

1. Secure defaults and auth maturity
2. App runtime hardening and config quality
3. Database and data-layer reliability
4. Better developer tooling and scaffolding
5. Stronger UI component system
6. Observability, jobs, and deployment readiness
7. Plugin architecture and extensibility

---

## Milestone 1: Runtime and config hardening

### Goal
Stabilize the core framework internals so apps boot predictably and safely in dev, test, and production.

### Deliverables
- central config loader with strict validation
- bootstrap lifecycle hooks: startup, shutdown, migrations, preflight checks
- structured environment profiles for development/test/production
- health and readiness endpoints
- graceful shutdown support
- improved error handling and request context tracing

### Key tasks
- add app startup sequence and lifecycle contract
- validate required environment variables at boot
- add config helpers for secret, DB, auth, and app settings
- add runtime diagnostics and health routes
- add tests for config validation and startup behavior

### Outcome
The framework behaves like a real app runtime, not just a collection of helpers.

---

## Milestone 2: Authentication and authorization maturity

### Goal
Turn auth from a demo flow into a secure, extensible system.

### Deliverables
- role-aware auth middleware
- permission checks and access control helpers
- session and token policy configuration
- OAuth/OIDC support
- MFA and passwordless authentication options
- audit logging for auth events and changes

### Key tasks
- add user roles and permissions model
- add route-level access policy helpers
- create RBAC middleware and example admin routes
- add JWT support for API auth
- add secure session expiration and rotation patterns
- add tests for unauthorized access, login edge cases, and expired sessions

### Outcome
Applications can ship secure auth with minimal custom code.

---

## Milestone 3: Persistence and data layer

### Goal
Make database usage easy, safe, and scalable beyond simple SQLite examples.

### Deliverables
- repository layer and query helpers
- DB driver abstraction for SQLite, Postgres, and MySQL patterns
- transaction helpers and migration health checks
- soft-delete and auditing conventions
- model validation hooks and serialization support

### Key tasks
- create typed repository patterns for core entities
- add transaction wrappers and DB error normalization
- improve migration management and verification tooling
- define generated model conventions for timestamps and audit fields
- add test coverage for query and transaction behavior

### Outcome
Applications can manage data with a consistent, production-safe abstraction.

---

## Milestone 4: Validation, forms, and request contracts

### Goal
Make request validation reliable and ergonomic across forms and APIs.

### Deliverables
- validator improvements with reusable rule sets
- sanitization rules and strict parsing helpers
- typed form binding for HTML and JSON requests
- structured validation output for UI and APIs
- rich examples for CRUD flows and form pages

### Key tasks
- support nested validation and better error composition
- improve form parsing for common data types
- add common rule sets for email, password, slug, UUID, and numeric fields
- add reusable API request validation patterns
- create example login/signup/admin forms using the framework primitives

### Outcome
Developers can build safe input flows without custom glue code.

---

## Milestone 5: UI system and component design

### Goal
Turn the UI layer into a coherent, reusable design system.

### Deliverables
- reusable primitives: buttons, inputs, cards, tables, modals, alerts
- layout and nav helpers
- page composition patterns for marketing, dashboard, and admin views
- HTMX-first patterns with progressive enhancement
- consistent styling conventions and design tokens

### Key tasks
- expand the UI package beyond primitives
- create starter dashboard and admin template sets
- add form helper components and error presentation patterns
- document page composition architectures
- add component usage examples for common workflows

### Outcome
The framework makes it straightforward to build polished, maintainable UIs.

---

## Milestone 6: Jobs, events, and async workflows

### Goal
Support real app workflows without blocking user requests.

### Deliverables
- improved event bus with retries and backoff
- background job abstraction
- scheduler support and worker health checks
- job observability and failure tracking
- queue-ready patterns for async processing

### Key tasks
- add event retry policies and dead-letter handling
- create job runner primitives
- add metrics around event drops and worker health
- support scheduled or recurring tasks
- document patterns for email, notifications, and imports

### Outcome
Applications can handle async work cleanly and reliably.

---

## Milestone 7: Observability and deployment readiness

### Goal
Make the framework production-friendly from the first deploy.

### Deliverables
- structured logging with request correlation IDs
- metrics and tracing hooks
- deployment profiles and environment separation
- release guidance and CI/CD examples
- operational dashboards and alerts patterns

### Key tasks
- add request-id propagation throughout app and logs
- add middleware timers and metrics instrumentation
- document production startup and deployment settings
- add graceful failure and recovery patterns
- create a production checklist for deployment

### Outcome
Teams can confidently ship the framework into staging and production.

---

## Milestone 8: Plugin ecosystem and app composition

### Goal
Create a framework-level extension model for real-world products.

### Deliverables
- plugin registration lifecycle
- plugin config injection
- example plugins for auth, admin, billing, analytics, and storage
- extension contracts for routes, views, middleware, and jobs
- compatibility and versioning strategy

### Key tasks
- define plugin interfaces and app registration patterns
- create example plugin modules in the bootstrap app
- add docs for plugin lifecycle and dependencies
- ensure plugin isolation and safe upgrade path
- add tests around plugin behavior and conflicts

### Outcome
Superkit becomes a framework people can extend without rewriting core behavior.

---

## Immediate backlog

### High priority
- runtime lifecycle and startup hardening
- auth roles and permission model
- config validation and security defaults
- repository and migration improvements
- test coverage for core auth and runtime flows

### Medium priority
- UI component library expansion
- validation and form helpers
- jobs and event retry infrastructure
- observability and logging improvements
- plugin contract design

### Lower priority but valuable
- CLI generators for routes, handlers, and models
- starter project templates
- marketing/admin dashboard examples
- docs site and reference app

---

## Suggested execution order

1. Runtime + config hardening
2. Auth + RBAC
3. Database layer
4. Validation + forms
5. UI components
6. Jobs + events
7. Observability
8. Plugin ecosystem

---

## Definition of done for a Grade 5 framework

The framework should support:
- secure auth and authorization by default
- strong validation and typed request handling
- reusable UI and component systems
- database abstraction and migration safety
- async jobs and event-driven workflows
- observability and production deployment support
- a plugin architecture for ecosystem growth
- complete examples and polished docs

## Final note

The repo is already a strong starting point. The right move is not to chase feature breadth randomly, but to build a focused sequence of milestones that turns this from a starter framework into a serious app platform.
