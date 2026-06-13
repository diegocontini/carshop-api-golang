# 003 - N+1 em upsert de itens de pedido

## Sintoma

Mesmo formato do bug 002, mas em `order_items` no lugar de `car_images`. Criar ou atualizar um pedido com K itens fazia K+ round trips ao banco so para os itens, mais um para a comissao. Um pedido com varios itens gastava a maior parte do tempo de wall-clock em round trips ao Postgres.

## Causa raiz (C#)

`CarShopApi/src/Services/OrderService.cs` `ProcessOrderItems` (linha 106): mesmo padrao de `SaveChangesAsync()` por iteracao visto em `CarService.ProcessImages`. Um round trip por item, independente de o item ser novo ou estar sendo atualizado.

## Correcao (Go)

`src/service/order_service.go` `Update`:

- Um SELECT para os IDs de item existentes (`ListOrderItemIDsByOrderID`).
- Um `BulkUpdateOrderItems` usando `unnest(bigint[], bigint[], numeric[], numeric[])` - atualiza todos os itens modificados em um unico statement.
- Um `BulkInsertOrderItems` via `:copyfrom` para os itens novos.
- Um `upsertCommission` (constante).

A sequencia inteira roda dentro de uma `pgx.BeginFunc` unica, entao uma falha no meio do caminho desfaz atomicamente tanto os updates de item quanto o update de comissao.

## Verificacao

Teste de integracao: `tests/integration/order_test.go` `TestOrderUpdateDoesNotDuplicateCommission` exercita o caminho de update de item ponta a ponta (o PUT substitui o `price` do item junto com o total do pedido). Combinado com `TestListOrdersByVendor` (que cria varios pedidos com um item cada), a suite cobre os caminhos de escrita com N itens.

Sem benchmark dedicado para "contar round trips" - os testes de integracao verificam correcao comportamental; o argumento de round trips e estrutural (um statement por tipo de escrita).
