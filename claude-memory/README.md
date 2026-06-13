# claude-memory

Implementation context for future Claude sessions. Each file answers a "future Claude reading this cold" question.

- [architecture.md](architecture.md) — layers, dependency direction, why no interfaces by default
- [auth.md](auth.md) — JWT scheme, role middleware, password hashing
- [domain-rules.md](domain-rules.md) — commission formula, seeding, contract preservations (e.g. "comission" misspelling)
- [decisions.md](decisions.md) — ADR-lite log of non-obvious choices

For C# bugs we fixed during the port, see `../bug-fixes-report/`.
For the HTTP contract, see `../docs/routes.md` (mirror of the C# `routes.md`).
