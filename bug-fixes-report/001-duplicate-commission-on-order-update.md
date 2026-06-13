# 001 - comissao-duplicada-em-update-de-pedido

## Sintoma

Cada `PUT /api/v1/order/{id}` inseria uma linha adicional em `vendor_comissions` em vez de atualizar a comissao ja existente daquele pedido. Apos N atualizacoes, o relatorio de comissoes do vendedor listava N linhas para o mesmo pedido, inflando os totais reportados.

## Causa raiz (C#)

`CarShopApi/src/Services/OrderService.cs` `UpdateAsync` chama `ProcessComission(existingOrder)` (linha 81). `ProcessComission` sempre constroi uma nova instancia de `VendorComission` e faz `AddAsync` - nunca procura uma comissao existente referenciada por `OrderId`.

```csharp
var comission = new VendorComission { ..., OrderId = order.Id ?? 0, ... };
await _db.VendorComissions.AddAsync(comission);
await _db.SaveChangesAsync();
```

Nao havia restricao no banco impedindo as duplicatas - `vendor_comissions.order_id` nao tinha indice unico.

## Correcao (Go)

Duas camadas:

1. **Schema** (`migrations/00001_init.sql`): `vendor_comissions.order_id` e `UNIQUE` com `REFERENCES orders(id) ON DELETE CASCADE`. O banco agora impoe o invariante.
2. **Service** (`src/service/order_service.go` `upsertCommission`): a escrita da comissao usa `UpsertComissionByOrder` (`src/infra/queries/comissions.sql`), que faz `INSERT ... ON CONFLICT (order_id) DO UPDATE SET ...`. Toda a escrita do pedido ocorre dentro de uma unica transacao `pgx.BeginFunc`.

O CASCADE em `order_id` tambem elimina o `_db.VendorComissions.RemoveRange(...)` manual que o `DeleteAsync` do C# precisava executar.

## Verificacao

Teste de integracao: `tests/integration/order_test.go` `TestOrderUpdateDoesNotDuplicateCommission`. Ele cria um pedido (3% de 100k -> 3000), faz PUT no mesmo pedido com total 200k e entao verifica:

- Exatamente 1 linha em `GET /api/v1/comission?vendorId=...`.
- O `comissionAmount` da linha equivale a `6000` (3% dos novos 200k).
