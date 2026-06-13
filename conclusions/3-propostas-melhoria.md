# Propostas de melhoria estrutural

Este documento descreve as melhorias estruturais aplicadas no port da API para Go (`carshop-api-golang/`). Atende ao Objetivo Especifico 5 do projeto basico. Cada secao apresenta a estrutura original em C# e a equivalente em Go, com a justificativa da mudanca.

## 1. Comparacao de arquitetura

| Aspecto | CarShopApi (C#) | carshop-api-golang (Go) | Justificativa |
|---|---|---|---|
| Framework HTTP | ASP.NET Core MVC | gin | Reduz boilerplate; tem middleware de auth mais explicito. |
| ORM / acesso a dados | EF Core (LINQ) | pgx + sqlc | Queries escritas em SQL puro; tipagem gerada a partir do schema; sem "tracking" implicito. |
| Migrations | EF Core (CLI separado) | goose (SQL puro) | Migrations sao arquivos `.sql` versionados; aplicaveis sem o framework. |
| Configuracao | `appsettings.json` + `IConfiguration` | struct `Settings` + env vars + `godotenv` | Validacao explicita de campos obrigatorios e tamanho minimo do JWT secret. |
| Tipo monetario | `double` (Car.Price) + `decimal` | `shopspring/decimal` em todos os campos | Elimina erro de ponto flutuante em qualquer calculo monetario. |
| Auth | JWT Bearer com claims pascal-case | JWT Bearer com claims lowercase + bcrypt | bcrypt aplicado em escrita, comparacao constant-time, senha removida das respostas. |
| Bootstrap admin | `if !exists then insert` (TOCTOU) | `INSERT ... ON CONFLICT DO NOTHING` | Idempotente e race-free no banco. |
| Foreign keys | Parciais (sem FK em vendor_comissions) | Completas com `ON DELETE CASCADE` | Integridade referencial garantida pelo banco. |
| Testes automatizados | Ausentes | 6 unit + 13 integration | Suite cobre contrato HTTP, regras de dominio e regressao da comissao duplicada. |
| Documentacao | Markdown estatico em `Docs/` | Markdown em `docs/` + `bug-fixes-report/` + `claude-memory/` | Documentacao segmentada por audiencia; tres categorias com proposito distinto. |
| Distribuicao | Imagem ASP.NET runtime (~250 MB) | Distroless static (~20 MB) | Reducao de superficie e tempo de pull. |

## 2. Estrutura de pastas

### C# (atual)

```
CarShopApi/
|- Program.cs
|- src/{Config,Controllers,Services,Models,Data,Docs}/
|- Migrations/                    (geradas pelo EF Core CLI)
'- Dockerfile, docker-compose.yml
```

### Go (proposto)

```
carshop-api-golang/
|- server/main.go                 entrypoint
|- src/
|  |- config/                     Settings + .env loader
|  |- controller/                 handlers gin, DTOs, middleware
|  |- service/                    regras de negocio
|  |- domain/                     entidades puras + invariantes
|  '- infra/{db,sqlc}/            pool pgx + queries geradas
|- queries/                       SQL de entrada do sqlc
|- migrations/                    SQL goose
|- tests/{unit,integration}/      suite completa
|- docs/                          contrato HTTP
|- claude-memory/                 contexto transversal (ingles)
|- bug-fixes-report/              um arquivo por bug corrigido
'- Dockerfile, docker-compose.yml, Makefile
```

Diferencas estruturais relevantes:

- A camada `domain/` nao existia no C#. Em Go, ela contem entidades puras e regras (ex.: `User.CalcCommission`) sem dependencia de banco ou HTTP, eliminando o acoplamento descrito em P-11.
- `queries/` e `migrations/` separam *o que* o codigo faz com o banco (SQL) de *como* (Go gerado pelo sqlc).
- `tests/`, `docs/`, `bug-fixes-report/` sao deliveraveis explicitos versionados junto ao codigo.

## 3. Acesso a dados

### C#

`AppDbContext` herda de `DbContext`. Services injetam o contexto e usam diretamente `_db.Cars.Include(...).Where(...)`. Updates dependem de "change tracking" implicito do EF Core.

### Go

- Schema definido em SQL puro em `migrations/00001_init.sql`.
- Queries escritas em SQL em `queries/*.sql` com anotacoes do sqlc (ex.: `-- name: ListUsers :many`).
- `sqlc generate` produz codigo Go tipado em `src/infra/sqlc/`. Cada query gera um metodo concreto com parametros e tipo de retorno.
- Services recebem `*pgxpool.Pool` e constroem `*sqlc.Queries`. Para transacoes, usam `pgx.BeginFunc(ctx, pool, func(tx) error { q := s.q.WithTx(tx); ... })`.

Beneficios:

- Toda query e auditavel em SQL antes da execucao.
- Bulk operations (`unnest(bigint[], text[])` para updates em lote, `COPY FROM` para inserts) eliminam o N+1 descrito em P-02 e P-03.
- O compilador Go falha caso o schema ou a query mudem incompativelmente.

## 4. Migrations

| Aspecto | C# (EF Core) | Go (goose) |
|---|---|---|
| Formato | Arquivos `.cs` gerados | Arquivos `.sql` escritos |
| Aplicacao em runtime | `context.Database.Migrate()` | `goose.UpContext(ctx, db, "migrations")` |
| Aplicacao manual | `dotnet ef database update` | `goose -dir migrations postgres "$URL" up` |
| Reversao | EF Core down | `goose down` |
| Versionamento | EF Core acoplado ao .NET | SQL independente da linguagem |

Migrations em SQL puro sao auditaveis sem ferramental .NET, podem ser executadas por DBA, e nao acoplam o esquema a uma versao especifica do framework.

## 5. Testes

A diferenca mais relevante. O projeto original nao possui testes; a versao em Go entrega uma suite executavel via `make test`.

### Cobertura por endpoint

| Endpoint | Testes em C# | Testes em Go | Tipo |
|---|---|---|---|
| `POST /api/v1/auth/token` | 0 | 3 (sucesso, 401, 400) | integration |
| `GET /api/v1/user` | 0 | 3 (401 sem token, 403 vendor, 200 admin, sem campo password) | integration |
| `POST /api/v1/user` | 0 | 2 (criacao + 409 em duplicata) | integration |
| `POST /api/v1/car` | 0 | 2 (403 vendor, 201 com imagens) | integration |
| `PUT /api/v1/car/{id}` | 0 | 1 (upsert misto: id existente + id novo) | integration |
| `POST /api/v1/order` | 0 | 1 (criacao + comissao calculada) | integration |
| `PUT /api/v1/order/{id}` | 0 | 1 (atualiza comissao sem duplicar - regressao P-01) | integration |
| `GET /api/v1/order?vendorId=N` | 0 | 1 (filtro por vendor) | integration |
| Regra `User.CalcCommission` | 0 | 5 (nulo, zero, negativo, inteiro, fracionario) | unit |
| Round-trip JWT | 0 | 2 (assinatura, secret errado) | unit |
| `bcrypt` hash + check | 0 | 1 | unit |
| `UserRole.IsValid` | 0 | 1 | unit |

Totais: **0 testes em C# vs 19 testes em Go (13 integration + 6 unit)**.

### Infraestrutura de testes

- `testcontainers-go` sobe `postgres:17-alpine` automaticamente antes da suite.
- `make test-integration` aplica migrations no container, constroi o router gin completo e executa cada teste via `httptest`.
- `SetupTest` faz `TRUNCATE ... RESTART IDENTITY CASCADE` e re-semeia o admin entre testes.
- Sem mocks de banco: todos os testes hit Postgres real, o mesmo da producao.

## 6. Seguranca e integridade

Cada um dos problemas de criticidade Alta do documento 2 e endereçado por uma mudanca estrutural na versao Go:

| Problema | Solucao estrutural |
|---|---|
| P-01 (comissao duplicada) | `UNIQUE (order_id)` em `vendor_comissions` + query `UpsertComissionByOrder` (`ON CONFLICT DO UPDATE`). |
| P-04 (senhas em texto claro) | `service.HashPassword` (bcrypt) em todo write; `UserResponse` DTO nao tem campo `Password`. |
| P-05 (TOCTOU no seed) | `UNIQUE (username)` + `INSERT ... ON CONFLICT (username) DO NOTHING`. |
| P-09 (ausencia de testes) | Suite integration + unit conforme tabela acima. |
| P-13 (JWT key default) | `config.Load` valida `len(JWT_SECRET) >= 32` no boot; ausencia da variavel falha a inicializacao. |

Os problemas de criticidade Media (P-02, P-03, P-06, P-07, P-08, P-11) sao endereçados respectivamente por bulk SQL via `unnest`, schema com `NUMERIC` para `cars.price`, FK com `ON DELETE CASCADE`, sentinel errors mapeados para HTTP no controller, e a camada `domain/` desacoplada.

## 7. Documentacao versionada

A versao Go separa documentacao por audiencia:

- `docs/routes.md` - contrato HTTP, fonte da verdade do comportamento externo. Espelha o C# e marca explicitamente as divergencias intencionais.
- `bug-fixes-report/` - um arquivo markdown por bug corrigido. Cada um contem sintoma, causa raiz em C# (com linha), correcao em Go (com referencia ao teste de verificacao).
- `claude-memory/` - contexto interno em ingles (arquitetura, decisoes, regras de dominio). Usado como base de conhecimento durante manutencao assistida.
- `CLAUDE.md` + `README.md` - entrada do repositorio em pt-BR com comandos uteis.

Nenhum desses artefatos existia no projeto original.

## 8. Sintese

A reestruturacao em Go nao se limita a uma troca de linguagem. Cada decisao estrutural - desde a separacao em `domain/`, passando pelo uso de sqlc, ate a suite de testes integrada - foi tomada para enderecar um problema concreto identificado na analise. O documento 4 (`4-avaliacao-migracao-go.md`) avalia se os ganhos compensaram o custo da migracao.
