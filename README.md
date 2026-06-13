# carshop-api-golang

Port em Go do `../CarShopApi` (C# / .NET 8). Mesmo contrato HTTP, qualidade melhor.

Stack: `gin` (HTTP) + `pgx/v5` + `sqlc` (banco, queries em `src/infra/queries/`) + `goose` (migrations) + JWT HS256 + bcrypt. Dinheiro em `shopspring/decimal`.

Documentacao complementar:
- `docs/routes.md` — contrato HTTP (fonte da verdade).
- `bug-fixes-report/` — um arquivo por bug do C# corrigido durante o port.
- `claude-memory/` — contexto interno (em ingles, para o Claude).
- `CLAUDE.md` — visao geral para quem vai abrir o repo pela primeira vez.

---

## Pre-requisitos

- **Go 1.26+**
- **Docker** — necessario para `docker-compose` e para a suite de integracao (`testcontainers-go`).
- **goose CLI** (opcional, so para rodar migrations fora do container):

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

- **sqlc CLI** (opcional, so para regerar queries):

  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```

---

## Variaveis de ambiente

Copie `.env.example` para `.env` e ajuste. O `make run` carrega `.env` automaticamente; em producao, defina as variaveis pelo orquestrador (Docker, Kubernetes, etc.).

| Variavel | Default | Descricao |
|---|---|---|
| `PORT` | `8080` | Porta HTTP. |
| `DATABASE_URL` | — | DSN do Postgres (`postgres://user:pass@host:port/db?sslmode=disable`). |
| `JWT_SECRET` | — | Chave HS256. Precisa ter pelo menos 32 caracteres. |
| `JWT_ISSUER` | — | Claim `iss`. |
| `JWT_AUDIENCE` | — | Claim `aud`. |
| `JWT_EXPIRY_MIN` | `60` | Validade do token em minutos. |
| `SUPERUSER_USERNAME` | — | Username do admin semeado no boot. |
| `SUPERUSER_PASSWORD` | — | Senha em texto claro (sera hasheada com bcrypt). |
| `SUPERUSER_EMAIL` | — | Email do admin semeado. |

---

## Como rodar

### Dev local

Precisa de um Postgres rodando (na sua maquina ou via compose). Com `.env` configurado:

```bash
make run
```

O `main` executa, na ordem: carrega config → abre o pool → roda `goose up` → semeia o admin (idempotente) → sobe o gin.

### Dev em Docker (stack completa)

Sobe Postgres + API juntos:

```bash
make docker-up      # docker compose up --build -d
make docker-down    # docker compose down
```

A API fica em `http://localhost:5001` (mapeado de `8080` do container). O Postgres em `localhost:5432`.

### Producao

Build da imagem (distroless, ~20MB, usuario nao-root):

```bash
docker build -t carshop-api .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@db:5432/carshopdb?sslmode=disable" \
  -e JWT_SECRET="<segredo-com-32-chars-ou-mais>" \
  -e JWT_ISSUER="CarShopApi" \
  -e JWT_AUDIENCE="CarShopApiClients" \
  -e JWT_EXPIRY_MIN=60 \
  -e SUPERUSER_USERNAME=admin \
  -e SUPERUSER_PASSWORD="<senha-forte>" \
  -e SUPERUSER_EMAIL="admin@example.com" \
  carshop-api
```

Em producao, as migrations sao aplicadas automaticamente no boot a partir do binario — nao precisa de container de migration separado.

---

## Testes

```bash
make test              # unit + integration (precisa de Docker)
make test-unit         # so unit (pacotes em src/ e tests/unit/)
make test-integration  # so integration (sobe um Postgres via testcontainers)
```

A suite de integracao sobe um `postgres:17-alpine` real via `testcontainers-go`, roda as migrations e exercita o gin completo via `httptest`. Cada teste comeca com tabelas truncadas e o admin re-semeado.

---

## Migrations (goose)

```bash
# Criar uma nova migration
make migrate-create name=add_index_orders
# Gera migrations/NNNN_add_index_orders.sql com os blocos +goose Up/Down

# Aplicar todas as pendentes
make migrate-up

# Desfazer a ultima
make migrate-down

# Ver o status
make migrate-status
```

Os comandos acima precisam de `DATABASE_URL` exportado e do binario do `goose` no `PATH`. Para rodar dentro do container, use `docker compose exec api goose -dir migrations ...`.

> Em producao, `migrate-up` nao precisa ser chamado manualmente — `server/main.go` aplica tudo que esta pendente quando o processo sobe.

---

## Geracao de codigo (sqlc)

Depois de editar arquivos em `src/infra/queries/`:

```bash
make sqlc            # roda `sqlc generate`
```

O codigo gerado fica em `src/infra/sqlc/`. Nao edite a mao — qualquer mudanca de tipo passa pelos arquivos `.sql` e por `sqlc.yaml`.

---

## Outros comandos uteis

```bash
make build      # compila um binario em ./bin/server
make fmt        # go fmt ./...
make lint       # golangci-lint run
make tidy       # go mod tidy
```
