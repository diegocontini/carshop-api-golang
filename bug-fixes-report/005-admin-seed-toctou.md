# 005 - seed-de-admin-toctou

## Sintoma

Em um banco limpo, dois processos da API subindo em paralelo conseguiam inserir o usuario admin os dois, deixando duas linhas com `username = 'admin'` (o schema do C# nao tinha indice unico). Logins subsequentes ficavam dependentes da ordem que o `FirstOrDefaultAsync` retornava.

## Causa raiz (C#)

`CarShopApi/Program.cs:39-58`:

```csharp
var hasAdmin = context.Users.Any(u => u.Username == "admin");
if (!hasAdmin)
{
    // ... build superUser ...
    context.Users.Add(superUser);
    context.SaveChanges();
}
```

Dois problemas combinados:
1. Janela de check-then-act entre o `Any(...)` e o `Add+SaveChanges` - TOCTOU classico.
2. Sem `UNIQUE (username)` no schema, entao o banco nao recusa o segundo insert.

## Correcao (Go)

Duas camadas:

1. **Schema** (`migrations/00001_init.sql`): `users.username TEXT NOT NULL UNIQUE`.
2. **Query** (`queries/users.sql` `SeedAdmin`):

```sql
INSERT INTO users (username, password, email, comission_per_sale_in_percent, role)
VALUES ($1, $2, $3, $4, 'admin')
ON CONFLICT (username) DO NOTHING;
```

Um unico statement SQL, executado uma vez por boot a partir do `server/main.go`. O banco resolve qualquer race deterministicamente; o segundo chamador recebe um no-op.

## Verificacao

- `tests/integration/harness_test.go` `SetupTest` re-executa `SeedAdmin` entre todos os testes contra uma tabela `users` recem-truncada. Uma chamada repetida contra um admin ja existente e um no-op e a suite continua verde - prova da idempotencia.
- Manualmente: `for i in 1 2 3; do make run & done` contra um banco vazio deixava 3 linhas de admin no C#. Com o novo schema, apenas uma linha existe.
