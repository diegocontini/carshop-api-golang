# 007 - vendor_comissions-sem-fk-para-orders

## Sintoma

Deletar um pedido nao removia as comissoes associadas automaticamente. O `OrderService.DeleteAsync` do C# tinha que fazer `_db.VendorComissions.RemoveRange(_db.VendorComissions.Where(c => c.OrderId == id))` manualmente. Se uma rota futura excluisse um pedido fora desse caminho, ficavam linhas orfas de `vendor_comissions` apontando para `order_id` inexistentes - corrompendo silenciosamente os relatorios de comissao.

## Causa raiz (C#)

`CarShopApi/src/Data/Mappings/VendorComissionMapping.cs` configurava so a tabela e a primary key, sem FK para `orders`. A migration gerada (`Migrations/20251110012750_InitialMigration.cs`) confirma a ausencia de qualquer foreign key em `vendor_comissions`:

```csharp
constraints: table => { table.PrimaryKey("PK_vendor_comissions", x => x.Id); }
```

`OrderService.DeleteAsync` tentou compensar fazendo a limpeza no codigo, mas isso so cobre o caminho atual.

## Correcao (Go)

`migrations/00001_init.sql`:

```sql
order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
```

A FK garante:
- Linhas orfas sao impossiveis a partir desse ponto.
- Deletar um pedido remove a comissao automaticamente - nenhum servico precisa lembrar de limpar.
- Combinada com o `UNIQUE` em `order_id` (ver bug-fixes-report/001), uma comissao por pedido, refencial e idempotente.

`DeleteOrder` em `src/service/order_service.go` agora e um simples `DELETE FROM orders WHERE id = $1` - o resto se resolve pelo CASCADE.

## Verificacao

A invariante e exercitada implicitamente pela suite de integracao: a `SetupTest` truncates `orders` com `CASCADE`, deixando `vendor_comissions` vazia entre testes. Se a FK CASCADE estivesse faltando, o TRUNCATE explodiria com `cannot truncate a table referenced in a foreign key constraint` (a `vendor_comissions` ficaria intocada). A suite passar verde e prova da configuracao.
