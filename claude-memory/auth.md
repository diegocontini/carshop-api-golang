# Auth

## JWT

- Algorithm: HS256, secret from `JWT_SECRET` env (>= 32 chars enforced at boot).
- Issuer + Audience: `JWT_ISSUER` / `JWT_AUDIENCE`, validated on every parse.
- Lifetime: `JWT_EXPIRY_MIN` minutes (default 60).
- Claims: `sub` = username, `role` = `"admin"` | `"vendor"` (lowercase), plus `jti` (uuid), `iat`, `exp`.
- The parser hard-codes `WithValidMethods([]string{"HS256"})` so `alg=none` and RS256-swap attacks are rejected up front.

The C# version embedded the role as PascalCase (`"Admin"`/`"Vendor"`). The Go port uses lowercase in the JWT because lowercase is already the DB representation. Clients re-login after the migration, so no token portability is needed.

## Role middleware

`middleware.RequireAuth(jwt)` pulls `Authorization: Bearer <token>`, parses it, and stuffs `sub` + `role` into the gin context. `middleware.RequireRoles(allowed...)` is a separate handler that gates by role and must come AFTER `RequireAuth`. Composition example in `controller/router.go`:

```go
v1.Group("/user", auth, adminOnly).Use(...)
```

Both 401 (no/invalid token) and 403 (wrong role) return `{"message": "..."}`.

## Passwords

- Hashed with `bcrypt.DefaultCost` on every write (Create, Update, SeedAdmin).
- Login: `service.CheckPassword(hash, plain)`.
- `Authenticate` returns `domain.ErrNotFound` for BOTH "unknown user" and "wrong password" so callers cannot enumerate usernames via timing or error class.
- The `password` field is OMITTED from every `/api/v1/user` GET response (single + list). The DTO `UserResponse` has no `Password` field at all - it's not a `json:"-"` trick, the field literally does not exist on the type. Intentional contract drift from the C# version where the plaintext password was echoed back.

## Wire representation of role

DB stores `'admin'`/`'vendor'` (TEXT + CHECK constraint). Domain uses `domain.UserRole` (typed string with the same values). HTTP wire uses PascalCase `"Admin"`/`"Vendor"` because the original C# clients expect it - translated in `controller/dto/user.go` via `RoleToDTO` / `RoleFromDTO`.

This dual representation is the load-bearing weirdness in user CRUD - if a test starts failing on role values, this is where to look.
