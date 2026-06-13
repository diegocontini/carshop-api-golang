# carshop-api-golang

Port em Go do `../CarShopApi` (C# / .NET 8). Mesmo contrato HTTP, qualidade melhor.

## Stack
- HTTP: `gin`
- Banco: `pgx/v5` + `sqlc` (queries em `/queries`, código gerado em `src/infra/sqlc`)
- Migrations: `goose` (SQL em `/migrations`, aplicado no boot a partir do `main`)
- Auth: JWT HS256 + senhas com bcrypt
- Dinheiro: `shopspring/decimal` (nada de `float` para valores monetarios)

## Layout
```
server/           ponto de entrada (estilo cmd)
src/config/       struct Settings carregada do env
src/controller/   handlers gin, DTOs, middleware
src/service/      regras de negocio (um arquivo por agregado)
src/domain/       entidades puras + invariantes
src/infra/db/     pool pgx + runner do goose
src/infra/sqlc/   queries geradas (NAO editar a mao)
migrations/       arquivos .sql do goose
queries/          arquivos .sql de entrada do sqlc
tests/            integration (testcontainers) + unit
docs/             contrato HTTP (routes.md)
claude-memory/    contexto transversal (leia antes de qualquer tarefa nao trivial)
bug-fixes-report/ um arquivo por bug do C# que foi corrigido no port
```

Direcao de dependencia: `controller -> service -> domain` e `service -> infra/sqlc`. `domain` nao importa nada interno.

## Comandos
- `make run` — dev local (precisa de `.env` + Postgres rodando)
- `make test` — unit + integration
- `make sqlc` — regerar codigo das queries (depois de editar `/queries`)
- `make docker-up` / `make docker-down` — stack completa via compose

## Leia antes de trabalhar
- `claude-memory/README.md` — indice (em ingles, contexto interno)
- `docs/routes.md` — contrato HTTP (fonte da verdade)
- `bug-fixes-report/README.md` — o que estava quebrado no C# e como foi corrigido
