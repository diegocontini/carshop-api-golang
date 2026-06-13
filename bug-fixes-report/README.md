# bug-fixes-report

One file per bug or smell present in the C# `CarShopApi` that the Go port fixes. Each file follows the same template:

```
# NNN — <slug>

## Symptom
What goes wrong, from the user's point of view.

## Root cause (C#)
File + line reference into `CarShopApi/`, brief explanation.

## Fix (Go)
What the Go code does instead, with file references into `carshop-api-golang/`.

## Verification
How we know the fix works (integration test name, manual repro, etc.).
```

Index:

- [001-duplicate-commission-on-order-update.md](001-duplicate-commission-on-order-update.md)
- [002-n-plus-one-on-image-upsert.md](002-n-plus-one-on-image-upsert.md)
- [003-n-plus-one-on-order-item-upsert.md](003-n-plus-one-on-order-item-upsert.md)
- [004-plaintext-passwords.md](004-plaintext-passwords.md)
- [005-admin-seed-toctou.md](005-admin-seed-toctou.md)
- [006-money-as-float.md](006-money-as-float.md)
- [007-vendor-commission-no-fk.md](007-vendor-commission-no-fk.md)
