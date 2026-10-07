# User Taste Learnings

- Communicates in Bahasa Indonesia (Indonesian) — respond in Indonesian unless they switch to English. Comfortably code-switches to English mid-session; follow the language of the most recent user message. Confidence: 0.9

- Strongly prefers command-line ergonomics: one-liner entry points (Makefile targets, PowerShell wrappers, `docker compose up`) over per-step instructions. When setting up a project, ship a `Makefile` + OS-portable script (`scripts/dev.ps1` on Windows) so common workflows (up/down/db/test/stress) are single commands. Confidence: 0.85

- README/docs should be written for a fresh developer joining the project, not as a post-hoc description of what was built. Include: prerequisites with install hints, a 3-command TL;DR, environment-variable table, daily-workflow table, manual fallback for users without the one-liner tooling. Confidence: 0.85

- Project layout preference: separate `backend/` and `frontend/` directories at the repo root (not monorepo-style nested under one `app/`), each self-contained with its own `Dockerfile`, plus a top-level `docker-compose.yaml`, `Makefile`, `README.md`, `ARCHITECTURE.md`. Confidence: 0.7