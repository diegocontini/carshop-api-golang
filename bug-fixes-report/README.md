# bug-fixes-report

Um arquivo por bug ou cheiro de codigo presente no `CarShopApi` (C#) que o port em Go corrigiu. Cada arquivo segue o mesmo template:

```
# NNN — <slug>

## Sintoma
O que da errado do ponto de vista do usuario.

## Causa raiz (C#)
Referencia para o arquivo + linha em `CarShopApi/`, explicacao breve.

## Correcao (Go)
O que o codigo em Go faz no lugar, com referencias para arquivos em `carshop-api-golang/`.

## Verificacao
Como sabemos que a correcao funciona (nome do teste de integracao, repro manual, etc.).
```

Indice:

- [001-duplicate-commission-on-order-update.md](001-duplicate-commission-on-order-update.md)
- [002-n-plus-one-on-image-upsert.md](002-n-plus-one-on-image-upsert.md)
- [003-n-plus-one-on-order-item-upsert.md](003-n-plus-one-on-order-item-upsert.md)
- [004-plaintext-passwords.md](004-plaintext-passwords.md)
- [005-admin-seed-toctou.md](005-admin-seed-toctou.md)
- [006-money-as-float.md](006-money-as-float.md)
- [007-vendor-commission-no-fk.md](007-vendor-commission-no-fk.md)
