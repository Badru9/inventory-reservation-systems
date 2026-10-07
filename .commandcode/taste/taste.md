# User Taste Learnings

- Communicates in Bahasa Indonesia (Indonesian) — respond in Indonesian unless they switch to English. Comfortably code-switches to English mid-session; follow the language of the most recent user message. Confidence: 0.9

- Prefers documentation over wrapper tooling: for project onboarding, put all commands directly in `README.md` (Docker / native paths, separate cmd / PowerShell / bash snippets for Windows variants) rather than shipping a `Makefile` + `scripts/dev.ps1` wrapper layer. Explicitly reversed an earlier Makefile + dev.ps1 approach in favor of README docs ("revert the makefile, and delete it. i think its better to just give the documentation for backend and frontend on README.md"). Confidence: 0.9

- README/docs should be written for a fresh developer joining the project, not as a post-hoc description of what was built. Include: prerequisites with install hints, a 3-command TL;DR, environment-variable table, daily-workflow table, manual fallback for users without the one-liner tooling. Confidence: 0.85

- Project layout preference: separate `backend/` and `frontend/` directories at the repo root (not monorepo-style nested under one `app/`), each self-contained with its own `Dockerfile`, plus a top-level `docker-compose.yaml`, `Makefile`, `README.md`, `ARCHITECTURE.md`. Confidence: 0.7