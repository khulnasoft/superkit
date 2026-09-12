# AGENTS.md

This repository is the Superkit Go framework plus a sample bootstrap app. Use the repo docs and build scripts as the source of truth.

## Quick start

- Read [README.md](README.md) for the public product overview and setup guidance.
- Use [Makefile](Makefile) for root-level checks and app commands.
- Use [bootstrap/Makefile](bootstrap/Makefile) for the example app workflow.

## Repo structure

- Root packages are framework-level utilities: `kit/`, `ui/`, `validate/`, `view/`, `db/`, `event/`.
- The sample app lives under [bootstrap/](bootstrap/). It contains the app, routes, views, migrations, and local tooling.
- Template files live under [bootstrap/app/views](bootstrap/app/views) and use Templ. Generated Go files are checked in and should be refreshed with the app generate command after template edits.

## Commands

Run the smallest relevant command for the change:

- Root unit tests: `make test`
- Bootstrap app tests: `make app-test`
- Generate Templ output: `make app-generate`
- Production build of the sample app: `make app-build`
- Development run for bootstrap app: `make app-dev`
- License compliance check: `make check-licenses`

## Conventions

- This project is Go-first. Prefer idiomatic Go, small focused packages, and tests next to the code they validate.
- When editing template components, update the `.templ` source and regenerate the corresponding `_templ.go` files instead of hand-editing generated output.
- Keep framework code and bootstrap app behavior intentionally separate: framework packages in the root are reusable, while the bootstrap app demonstrates usage.
- Avoid broad refactors or large dependency churn without a clear reason; the repo favors lightweight utilities and minimal ceremony.

## Working style for AI agents

- Before editing, identify whether the code belongs to the framework root or the bootstrap example app.
- For user-facing web changes, inspect nearby handlers, routes, and views together so the change is coherent.
- Prefer targeted tests and focused validation over broad repository-wide runs.
- If you need the full workflow for the demo app, use the scripts in [bootstrap/Makefile](bootstrap/Makefile) and not ad hoc commands.

## Useful references

- [README.md](README.md)
- [Makefile](Makefile)
- [bootstrap/Makefile](bootstrap/Makefile)
- [bootstrap/app/routes.go](bootstrap/app/routes.go)
- [bootstrap/app/handlers](bootstrap/app/handlers)
- [bootstrap/app/views](bootstrap/app/views)
