# 006 - dinheiro-como-float

## Sintoma

`Car.Price` podia perder precisao em operacoes aritmeticas. Um carro com preco `99999.99` armazenado como `double` faria round trip seguro para esse valor especifico, mas qualquer calculo posterior (soma de precos, conversao de moeda, aplicacao de desconto) podia divergir.

O bug nunca se manifestou em producao porque o codigo nunca *calculava* com `Car.Price` - era so display. E um bug latente, corrigido preventivamente como parte do port.

## Causa raiz (C#)

`CarShopApi/src/Models/Car.cs:14`:

```csharp
public double Price { get; set; }
```

Mapeado para Postgres `double precision`. O EF Core respeita qualquer tipo C# que voce passar.

Os outros campos monetarios (`Order.Total`, `OrderItem.Price/Discount`, valores de comissao) ja eram `decimal` em C# - esses traduzem direto para `NUMERIC` e nao precisaram mudar.

## Correcao (Go)

- `migrations/00001_init.sql`: `cars.price` e `NUMERIC NOT NULL`. Todas as colunas monetarias do schema sao `NUMERIC`.
- `src/domain/domain.go`: todo campo monetario tipado e `decimal.Decimal` (`github.com/shopspring/decimal`).
- `sqlc.yaml`: type overrides mapeiam `pg_catalog.numeric` -> `shopspring/decimal`. O codigo gerado usa o mesmo tipo ponta a ponta.
- DTOs serializam `decimal.Decimal` como numero JSON (comportamento padrao da lib), preservando o formato de wire que os clientes C# enviam.

## Verificacao

- `tests/unit/domain_test.go` `TestUserCalcCommission` exercita o caminho sensivel a precisao: `3% * 12345.67 = 370.3701`. Um calculo em `float64` arredondaria para `370.37009999...` e quebraria a igualdade.
- Testes de integracao exercitam o tipo da coluna no schema implicitamente - qualquer drift de arredondamento apareceria quando o `6000` esperado do teste de regressao fosse comparado com um valor levemente diferente.
