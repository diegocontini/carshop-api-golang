# Documento de analise da API

Este documento mapeia a API CarShopApi (C# / .NET 8) que foi escolhida como objeto de estudo da disciplina de Manutencao de Software. Atende ao Objetivo Especifico 1 do projeto basico ("Mapear as funcionalidades principais da API") e e a base sobre a qual os demais documentos sao construidos.

## 1. Identificacao do sistema

- **Repositorio**: `CarShopApi/` (sibling de `carshop-api-golang/`).
- **Stack**: .NET 8, ASP.NET Core, Entity Framework Core 9, PostgreSQL via Npgsql, autenticacao JWT Bearer, documentacao via Swashbuckle (Swagger).
- **Persistencia**: PostgreSQL 17 (configurado em `docker-compose.yml`).
- **Distribuicao**: imagem Docker baseada em `mcr.microsoft.com/dotnet/aspnet:8.0`.

## 2. Estrutura de pastas

```
CarShopApi/
|- Program.cs
|- src/
|  |- Config/             DependencyInjector.cs
|  |- Controllers/        AuthController, CarController, ComissionController,
|  |                      OrderController, UserController, Dtos/
|  |- Services/           CarService, ComissionService, OrderService,
|  |                      UserService, JwtService
|  |- Models/             Car, CarImage, Order, OrderItem, User, VendorComission
|  |- Data/               CarShopDbContext, Mappings/
|  '- Docs/               routes.md
|- Migrations/            EF Core (geradas)
|- Dockerfile
'- docker-compose.yml
```

A organizacao segue o estilo padrao MVC do ASP.NET: controllers chamam services, services usam o `AppDbContext` (EF Core) diretamente. Nao ha camada de repositorio explicita.

## 3. Funcionalidades principais

A API expoe um CRUD completo sobre quatro agregados de dominio, alem do fluxo de autenticacao. Em alto nivel:

| Agregado | Operacoes | Caminho base |
|---|---|---|
| Autenticacao | login com username/password, gera JWT | `POST /api/v1/auth/token` |
| User | CRUD restrito a Admin | `/api/v1/user` |
| Car | CRUD, imagens aninhadas | `/api/v1/car` |
| Order | CRUD, gera comissao automatica | `/api/v1/order` |
| VendorComission | CRUD manual, filtro por vendor | `/api/v1/comission` |

Endpoints auxiliares: `GET /` (redireciona para `/swagger`) e `GET /health`.

## 4. Modelo de dados

Seis tabelas. Diagrama de relacionamentos:

```
users (1) ----< orders (1) ----< order_items (N) >---- (1) cars (1) ----< car_images (N)
   |               |
   |               '----< vendor_comissions (1)
   '------------------------(referencia logica via vendor_id)
```

Campos relevantes:

- `users.role`: enum (`Admin` = 0, `Vendor` = 1) persistido como inteiro.
- `users.comission_per_sale_in_percent`: percentual base usado no calculo de comissao.
- `cars.price`: tipo `double` (ponto-flutuante).
- `orders.total`, `order_items.price/discount`, `vendor_comissions.*`: tipo `decimal`.

Observacao: relacionamento `vendor_comissions.order_id -> orders.id` existe so na logica do service. Nao havia foreign key no schema gerado pela migration EF Core.

## 5. Fluxo principal: criacao de pedido

1. Cliente envia `POST /api/v1/order` com `CreateOrUpdateOrderDto` contendo `vendorId`, `total` e a lista de `items`.
2. `OrderController.Create` chama `OrderService.CreateAsync`.
3. `OrderService` abre uma transacao, insere a `Order`, processa os `OrderItem`s e processa a `VendorComission`.
4. `ProcessComission` carrega o usuario (vendor), calcula `comissionAmount = (percent / 100) * total` e insere uma nova linha em `vendor_comissions`.
5. Em caso de sucesso, commit. Em falha, rollback.

Esse fluxo e o mais complexo da API e concentra a maior parte dos problemas estruturais identificados no documento seguinte.

## 6. Autenticacao e autorizacao

- JWT Bearer assinado com HS256 (chave em `appsettings.json` por padrao).
- Claims emitidas pelo `JwtService`: `sub` (username), `role` (`Admin`/`Vendor`), `jti`.
- Gate de role aplicado por atributos `[Authorize(Roles = "Admin")]` / `[Authorize(Roles = "Admin,Vendor")]` nos controllers.
- Bootstrap em `Program.cs` cria automaticamente um superuser admin se nao existir.

## 7. Estrategia de execucao

A aplicacao roda em container. O `docker-compose.yml` sobe simultaneamente:

- `db`: PostgreSQL 17 com healthcheck.
- `api`: imagem do .NET 8 aspnet runtime, expondo a porta 8080.

As migrations sao aplicadas em runtime pelo `Program.cs` via `context.Database.Migrate()` na inicializacao.

## 8. Sintese

A CarShopApi atende ao requisito funcional - implementa o CRUD basico de um sistema de garagem de automoveis e oferece os fluxos de autenticacao e calculo de comissao. A organizacao em camadas (`Controllers` -> `Services` -> `Data`) e consistente com a documentacao ASP.NET, e o uso de EF Core simplifica a integracao com o banco.

No entanto, o sistema apresenta problemas de manutenibilidade, performance e seguranca que justificam a analise feita no documento 2 (`2-problemas-identificados.md`) e as propostas dos documentos 3, 4 e 5.
