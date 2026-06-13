# 002 - N+1 em upsert de imagens de carro

## Sintoma

Criar ou atualizar um carro com K imagens enviava K+ idas e voltas ao banco: uma por imagem, as vezes duas (um SELECT para checar se a linha existe, depois um UPDATE/INSERT). Para uma tela de detalhe de carro tipica enviando ~8 imagens, isso era ~16 round trips por requisicao.

## Causa raiz (C#)

`CarShopApi/src/Services/CarService.cs` `ProcessImages` (linha 90) percorre as imagens e dentro do loop chama `_db.SaveChangesAsync()` uma vez por imagem. O EF Core agrupa statements dentro de um unico `SaveChangesAsync`, mas nao entre chamadas - entao cada iteracao e um round trip separado:

```csharp
foreach (var imageDto in imageDtos)
{
    if (imageDto.Id.HasValue) {
        var existingImage = await _db.CarImages.FirstOrDefaultAsync(...);
        if (existingImage != null) {
            existingImage.Url = imageDto.Url;
            await _db.SaveChangesAsync();   // round trip
            continue;
        }
    }
    var newImage = new CarImage { ... };
    await _db.CarImages.AddAsync(newImage);
    await _db.SaveChangesAsync();           // round trip
}
```

## Correcao (Go)

`src/service/car_service.go` `Update`:

1. Um SELECT para os IDs de imagem existentes (`ListCarImageIDsByCarID`).
2. Classifica as imagens da requisicao em "lista de update" (ID existente bate) e "lista de insert" (sem ID ou ID desconhecido).
3. Um `BulkUpdateCarImages` usando `unnest(bigint[], text[])` para atualizar todas as linhas num unico statement.
4. Um `BulkInsertCarImages` (sqlc `:copyfrom`) para os inserts.

O total de round trips para K imagens fica em no maximo 3, independente de K. Mais um `ListCarImagesByCarIDs` final para recarregar o estado canonico da resposta - ainda constante.

`Create` e ainda mais simples: um unico `CopyFrom` para toda a lista de imagens.

## Verificacao

Teste de integracao: `tests/integration/car_test.go` `TestUpdateCarUpsertsImages` cria um carro com uma imagem, depois faz PUT no carro com `images: [{id: existente, url: "updated"}, {url: "new"}]`. A resposta contem exatamente 2 imagens, uma com a URL atualizada e outra recem-inserida - tudo em uma unica chamada HTTP suportada por uma unica transacao no banco.
