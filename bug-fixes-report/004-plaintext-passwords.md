# 004 - senhas-em-texto-claro

## Sintoma

As senhas de todos os usuarios (incluindo o admin semeado) eram armazenadas na coluna `users.password` em texto claro. `GET /api/v1/user` e `GET /api/v1/user/{id}` devolviam a senha em texto claro para qualquer chamador que passasse pelo gate de role Admin. Qualquer pessoa com acesso de leitura ao banco (backup, log, stacktrace de ORM) via a mesma coisa.

## Causa raiz (C#)

- `CarShopApi/Program.cs:50`: o seed do superuser grava `Password = adminPassword` direto.
- `CarShopApi/src/Services/UserService.cs:58`: `AuthenticateLogin` compara em texto claro com `u.Username == dto.Username && u.Password == dto.Password`.
- `User.Password` e uma property comum, sem nenhum atributo de JSON-ignore, entao e ecoada em toda leitura.

## Correcao (Go)

Tres mudancas:

1. **Hash na escrita** (`src/service/user_service.go`): `Create`, `Update` e `SeedAdmin` chamam `service.HashPassword` (bcrypt em `bcrypt.DefaultCost`) antes de persistir.
2. **Comparacao em tempo constante no login** (`src/service/user_service.go` `Authenticate`): `service.CheckPassword(row.Password, dtoPassword)` encapsula `bcrypt.CompareHashAndPassword`. Retorna `domain.ErrNotFound` tanto para usernames desconhecidos quanto para senha errada, entao o chamador nao consegue enumerar usernames pela classe do erro.
3. **Omite da resposta** (`src/controller/dto/user.go`): `UserResponse` nao tem o campo `Password` (nem mesmo como `json:"-"`). O conversor do DTO so copia o subset seguro.

## Verificacao

- `tests/unit/auth_test.go` `TestPasswordHashRoundTrip` cobre hash/check.
- `tests/integration/auth_user_test.go` `TestCreateUserOmitsPasswordInResponse` confirma que o corpo da resposta nao contem nem a string `"password"` nem o valor em texto claro.
- `TestListUsersDoesNotIncludePasswordHash` percorre cada usuario da resposta da lista e confirma que nao existe a chave `password`.
