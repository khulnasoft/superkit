---
applyTo: "bootstrap/**"
---

# Bootstrap app workflow

Use the bootstrap app as a separate project with its own task flow.

- Run commands from [bootstrap/Makefile](../../bootstrap/Makefile) instead of inventing ad hoc build steps.
- Prefer the smallest relevant command:
  - `cd bootstrap && make test`
  - `cd bootstrap && make dev`
  - `cd bootstrap && make build`
  - `cd bootstrap && make generate`
- Treat the app as a Go project built around the example server under [bootstrap/app](../../bootstrap/app).
- Keep UI changes coherent across route handlers, view files, and templates. If a page changes, inspect the matching route and view together.
- When editing Templ components, update the `.templ` source and regenerate the generated Go files with `make generate` (or the root `make app-generate`) rather than hand-editing generated output.
- App asset watchers and proxy tooling are defined in [bootstrap/Makefile](../../bootstrap/Makefile); use them for local dev rather than custom scripts.
- The root framework packages in [kit](../../kit), [ui](../../ui), [view](../../view), [validate](../../validate), [db](../../db), and [event](../../event) are reusable infrastructure. Keep bootstrap app behavior in the app folder unless the change is clearly framework-level.
- Prefer targeted tests for the changed behavior over broad repo-wide validation.
