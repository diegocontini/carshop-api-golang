# Lista de problemas de manutencao identificados

Este documento cataloga os problemas encontrados na analise da CarShopApi. Atende aos Objetivos Especificos 2, 3, 4 e 7 do projeto basico. Cada item segue o formato: descricao, localizacao no codigo, impacto e criticidade.

Criticidade: **Alta** (compromete dados ou seguranca), **Media** (compromete desempenho ou manutenibilidade), **Baixa** (organizacao ou estilo).

## P-01. Comissao duplicada em update de pedido

- **Descricao**: o metodo `OrderService.UpdateAsync` chama `ProcessComission(existingOrder)` que sempre insere uma nova linha em `vendor_comissions`. Apos N updates no mesmo pedido, existem N linhas de comissao apontando para o mesmo `order_id`, inflando relatorios.
- **Localizacao**: `CarShopApi/src/Services/OrderService.cs:81-90`.
- **Impacto**: corrompe dados de comissao do vendor.
- **Criticidade**: **Alta**.

## P-02. N+1 em upsert de imagens de carro

- **Descricao**: `ProcessImages` percorre as imagens e chama `SaveChangesAsync` dentro do loop. Cada imagem dispara um round trip ao banco. Para uma chamada com 8 imagens novas e 4 atualizadas, o custo e da ordem de 16-20 round trips.
- **Localizacao**: `CarShopApi/src/Services/CarService.cs:90-115`.
- **Impacto**: degradacao de desempenho proporcional ao numero de imagens.
- **Criticidade**: **Media**.

## P-03. N+1 em upsert de itens de pedido

- **Descricao**: mesmo padrao do P-02, em `OrderService.ProcessOrderItems`. Cada item exige um round trip independente.
- **Localizacao**: `CarShopApi/src/Services/OrderService.cs:106-135`.
- **Impacto**: pedidos com muitos itens degradam o tempo de resposta de forma linear.
- **Criticidade**: **Media**.

## P-04. Senhas armazenadas em texto claro

- **Descricao**: `User.Password` e gravado sem hash. `UserService.AuthenticateLogin` compara senha em texto claro com `u.Password == dto.Password`. Alem disso, `User.Password` nao tem `[JsonIgnore]`, entao a senha e ecoada em qualquer resposta de `GET /api/v1/user`.
- **Localizacao**: `CarShopApi/src/Services/UserService.cs:58`, `CarShopApi/src/Models/User.cs`, `CarShopApi/Program.cs:50`.
- **Impacto**: exposicao de credenciais via API, logs, backup ou stacktrace. Vazamento permanente em caso de dump do banco.
- **Criticidade**: **Alta**.

## P-05. TOCTOU no seed do superuser admin

- **Descricao**: o bootstrap em `Program.cs` faz `if (!context.Users.Any(u => u.Username == "admin")) { context.Users.Add(...); context.SaveChanges(); }`. Dois processos subindo em paralelo executam o check antes de qualquer um inserir, gerando dois usuarios admin. Nao ha constraint `UNIQUE (username)` no schema para impedir o segundo insert.
- **Localizacao**: `CarShopApi/Program.cs:39-58`, `CarShopApi/Migrations/20251110012750_InitialMigration.cs`.
- **Impacto**: estado inconsistente; logins subsequentes ficam dependentes da ordem de retorno do `FirstOrDefaultAsync`.
- **Criticidade**: **Alta**.

## P-06. Tipo de moeda inadequado em `Car.Price`

- **Descricao**: `Car.Price` e declarado como `double`, mapeado para `double precision` no Postgres. Os demais campos monetarios (`Order.Total`, `OrderItem.Price`, etc.) ja usam `decimal`. O sistema atual nao calcula sobre `Car.Price`, mas qualquer uso futuro em soma, conversao de moeda ou aplicacao de desconto introduz erro de ponto flutuante.
- **Localizacao**: `CarShopApi/src/Models/Car.cs:14`.
- **Impacto**: bug latente de precisao em calculos monetarios.
- **Criticidade**: **Media**.

## P-07. Falta de foreign key em `vendor_comissions.order_id`

- **Descricao**: `vendor_comissions` nao tem FK para `orders`. A integridade referencial e mantida manualmente em `OrderService.DeleteAsync` via `RemoveRange`. Qualquer delete fora desse caminho produz linhas orfas.
- **Localizacao**: `CarShopApi/src/Data/Mappings/VendorComissionMapping.cs`, `CarShopApi/src/Services/OrderService.cs:92-104`.
- **Impacto**: integridade referencial fragil; relatorios podem incluir comissoes de pedidos inexistentes.
- **Criticidade**: **Media**.

## P-08. Tratamento de erro generico no service

- **Descricao**: violacoes de regra de negocio sao reportadas via `throw new Exception("...")`. O controller captura com `catch (Exception ex)` e retorna 400 com a mensagem da excecao, perdendo a distincao entre 409 (conflito), 400 (input invalido) e 500 (erro interno).
- **Localizacao**: `CarShopApi/src/Services/UserService.cs:17`, `CarShopApi/src/Controllers/UserController.cs:46-54`.
- **Impacto**: erros 409 sao mascarados como 400; expoe mensagens internas; clientes nao conseguem reagir programaticamente ao tipo de erro.
- **Criticidade**: **Media**.

## P-09. Ausencia de testes automatizados

- **Descricao**: nao existem projetos de teste no `.sln`. Nao ha unit tests, integration tests, nem fixtures de banco. Qualquer mudanca depende de verificacao manual.
- **Localizacao**: ausencia em todo o repositorio.
- **Impacto**: alta probabilidade de regressao silenciosa em qualquer alteracao; impede refatoracao com seguranca.
- **Criticidade**: **Alta**.

## P-10. Mapeamentos EF Core vazios

- **Descricao**: arquivos `CarMapping.cs`, `OrderMapping.cs`, etc. apenas configuram `ToTable` e `HasKey`. Nao definem indices, FKs, constraints de coluna nem tipos especificos. A configuracao real do schema (tipos, defaults, unicidade) nasce com defaults do EF Core - nem sempre adequados ao Postgres.
- **Localizacao**: `CarShopApi/src/Data/Mappings/*.cs`.
- **Impacto**: schema fica subespecificado; problemas como P-05 e P-07 derivam dessa lacuna.
- **Criticidade**: **Baixa** (estrutural).

## P-11. Acoplamento direto entre controllers e DbContext atraves do service

- **Descricao**: services recebem o `AppDbContext` no construtor e usam diretamente as `DbSet<T>` e queries LINQ. Nao existe camada de repositorio nem abstracao da fonte de dados. Qualquer mudanca de ORM (por exemplo migrar para Dapper ou outro provedor) toca todos os services.
- **Localizacao**: `CarShopApi/src/Services/*.cs`.
- **Impacto**: dificulta evolucao tecnologica e testes (nao da para mockar a fonte de dados sem ferramentas adicionais).
- **Criticidade**: **Media**.

## P-12. Configuracao de banco com fallback obscuro

- **Descricao**: `DependencyInjector.Inject` tenta usar `ConnectionStrings:DefaultConnection`; se vazio, monta a connection string a partir de variaveis individuais (`DB_HOST`, `DB_PORT`, etc.). Sem documentacao clara sobre qual variavel tem prioridade ou e obrigatoria.
- **Localizacao**: `CarShopApi/src/Config/DependencyInjector.cs:14-30`.
- **Impacto**: falha de configuracao silenciosa; difficulta deploy.
- **Criticidade**: **Baixa**.

## P-13. JWT key default no `appsettings.json`

- **Descricao**: `appsettings.json` carrega como default a string `your-super-secret-key-that-should-be-at-least-32-characters-long`. Se o operador esquecer de sobrescrever em producao, a API sobe com chave conhecida publicamente.
- **Localizacao**: `CarShopApi/appsettings.json:13`.
- **Impacto**: tokens JWT podem ser forjados; equivale a ausencia de autenticacao.
- **Criticidade**: **Alta**.

## P-14. Documentacao do contrato HTTP nao versionada com o codigo

- **Descricao**: o arquivo `src/Docs/routes.md` documenta o contrato esperado, mas nao ha mecanismo automatico que falhe o build se o codigo divergir do documento.
- **Localizacao**: `CarShopApi/src/Docs/routes.md`.
- **Impacto**: documento envelhece silenciosamente.
- **Criticidade**: **Baixa**.

## Sintese e priorizacao

| # | Problema | Criticidade |
|---|---|---|
| P-01 | Comissao duplicada em update | Alta |
| P-04 | Senhas em texto claro | Alta |
| P-05 | TOCTOU no seed do admin | Alta |
| P-09 | Ausencia de testes | Alta |
| P-13 | JWT key default exposta | Alta |
| P-02 | N+1 em imagens | Media |
| P-03 | N+1 em itens de pedido | Media |
| P-06 | Money como double | Media |
| P-07 | Falta de FK em comissao | Media |
| P-08 | Tratamento de erro generico | Media |
| P-11 | Acoplamento service-DbContext | Media |
| P-10 | Mapeamentos EF Core vazios | Baixa |
| P-12 | Configuracao com fallback obscuro | Baixa |
| P-14 | Doc de rotas nao versionada | Baixa |

O documento 3 (`3-propostas-melhoria.md`) descreve como a versao em Go endereca cada um desses itens. O documento `bug-fixes-report/` no repositorio `carshop-api-golang/` traz a referencia tecnica detalhada de cada correcao.
