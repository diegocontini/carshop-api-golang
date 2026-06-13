# Avaliacao da migracao para Go

Este documento avalia a viabilidade, dificuldades e ganhos observados na migracao da CarShopApi (C#) para `carshop-api-golang` (Go). Atende ao Objetivo Especifico 6 do projeto basico.

## 1. Estrategia de migracao adotada

Foi feita uma reescrita preservando o contrato HTTP, nao um rewrite "big-bang". Premissa: clientes existentes continuam funcionando sem alteracao apos a troca da implementacao.

Etapas (mapeadas para os 14 commits do repositorio Go):

1. Scaffold do modulo Go (`utfpr.edu.br/carshop-api`).
2. Schema reescrito em SQL puro com correcoes de integridade.
3. Geracao de codigo via sqlc.
4. Camadas `config`, `domain`, `service`, `controller` reconstruidas uma a uma.
5. Suite de testes em paralelo com o desenvolvimento das features.
6. Wiring final e documentacao.

Cada commit deixa o repositorio compilavel e testavel, o que permitiu validacao incremental.

## 2. Dificuldades encontradas

### 2.1. Limitacao do `//go:embed` para migrations

O Go nao permite que `//go:embed` faca referencia a paths fora do diretorio do arquivo fonte. Como o projeto basico exige `migrations/` no nivel raiz, e o package `db` esta em `src/infra/db/`, foi necessario criar um arquivo `embed.go` no nivel raiz (package `carshop`) so para expor o `embed.FS`, que e injetado em `db.Migrate`. E uma concessao estrutural - duas locais para uma responsabilidade - mas necessaria para atender o layout.

### 2.2. Mapeamento de tipos monetarios no sqlc

Por padrao, sqlc mapeia `numeric` para `string`. Foi preciso configurar overrides em `sqlc.yaml`:

```yaml
overrides:
  - db_type: "pg_catalog.numeric"
    go_type:
      import: "github.com/shopspring/decimal"
      type: "Decimal"
```

Esse tipo de configuracao exige conhecimento do sqlc; a primeira execucao gerou codigo com `string` e quebrou a integracao com `shopspring/decimal`.

### 2.3. Convencao do campo `role` na fronteira

A API C# emite o role como string PascalCase (`"Admin"` / `"Vendor"`) por causa do `JsonStringEnumConverter` do `System.Text.Json`. No banco PostgreSQL, optou-se por armazenar lowercase (`'admin'` / `'vendor'`) com restricao `CHECK`. Foi necessario implementar um par de funcoes `RoleFromDTO` / `RoleToDTO` na camada de DTO para preservar o contrato externo. Adiciona ~30 linhas mas evita reescrever o cliente movel.

### 2.4. Decisao sobre o campo `password` em respostas

O contrato original incluia `password` em respostas de `GET /api/v1/user`. Manter esse comportamento implicaria expor o hash bcrypt em todas as listagens - praticamente o mesmo problema do plaintext. Optou-se por omitir o campo da resposta, documentando a divergencia em `docs/routes.md` e em `bug-fixes-report/004-plaintext-passwords.md`. E o unico ponto onde o contrato HTTP foi alterado.

### 2.5. Preservacao da grafia "comission"

O codigo C# original tem o erro de digitacao "Comission" (com um 'm' apenas) em todos os lugares: nome de tabela, nome de DTO, path de rota (`/api/v1/comission`), nome de campo JSON. Corrigir esse erro implicaria quebrar todos os clientes. Decidiu-se preservar a grafia errada em toda a fronteira externa e no banco, com nota explicita em `claude-memory/domain-rules.md`.

## 3. Vantagens observadas

### 3.1. Reducao da imagem Docker

- C#: `mcr.microsoft.com/dotnet/aspnet:8.0` como base. Tamanho final tipico ~250 MB.
- Go: `gcr.io/distroless/static-debian12:nonroot` com binario estatico. Tamanho final ~20 MB.

Reducao de ~92% no tamanho da imagem. Beneficios diretos: pull mais rapido em CI/CD, menor superficie de ataque (sem shell, sem package manager no runtime), menor consumo de espaco em registries.

### 3.2. Tempo de boot

O binario Go inicia em poucas centenas de milissegundos. O processo inteiro (load config -> abrir pool -> aplicar migrations -> semear admin -> subir gin) executa em menos de 1s contra um banco ja inicializado. Util para ambientes de CI e para sondas de readiness do Kubernetes.

### 3.3. SQL explicito reduz surpresas

EF Core gera queries em runtime e aplica otimizacoes baseadas em "change tracking". Algumas das ineficiencias identificadas (P-02, P-03) sao consequencia desse comportamento implicito - o desenvolvedor escreve `foreach + SaveChangesAsync` esperando batching e nao recebe.

sqlc obriga a escrever cada query em SQL antes do build. O que esta no codigo e o que executa. Bulk operations precisam ser modeladas explicitamente, o que dificulta introduzir N+1 sem perceber.

### 3.4. testcontainers-go simplifica testes de integracao

A suite roda em qualquer maquina com Docker. Nao depende de `docker-compose up` previo, nao precisa de banco "de teste" pre-instalado, e nao polui um banco compartilhado entre testes. Cada `go test` e independente.

### 3.5. Tipagem mais simples em interfaces externas

O Go padroniza a representacao de tempo (`time.Time`), numeros (`int64`, `int32`), e textos (`string`) sem niveis adicionais como `long?` ou `byte?`. A traducao para JSON e direta. Diferentemente do C#, onde foi necessario lidar com `JsonStringEnumConverter` e `ReferenceHandler.IgnoreCycles` para evitar serializar ciclos do EF Core.

### 3.6. Compilacao verifica o uso do banco

Como o sqlc gera tipos para cada query, qualquer mudanca de schema que afete um servico produz erro de compilacao imediato. Em EF Core, mudancas comparaveis podem aparecer so em runtime durante o `SaveChangesAsync`.

## 4. Melhorias notaveis

Quantitativamente, comparando o estado final em Go com o codigo original em C#:

| Metrica | C# (CarShopApi) | Go (carshop-api-golang) |
|---|---|---|
| Testes automatizados | 0 | 19 (6 unit + 13 integration) |
| Bugs criticos abertos (Alta) | 5 (P-01, P-04, P-05, P-09, P-13) | 0 |
| Bugs medios abertos (Media) | 6 (P-02, P-03, P-06, P-07, P-08, P-11) | 0 |
| Foreign keys no schema | 2 | 5 |
| Constraints de unicidade | 0 | 2 (`users.username`, `vendor_comissions.order_id`) |
| Round trips ao banco para criar pedido com K itens | K + ~3 | 3 |
| Round trips ao banco para atualizar carro com K imagens | ~2K + 2 | 3 |
| Tamanho da imagem Docker | ~250 MB | ~20 MB |
| Documentacao versionada (arquivos `.md` no repo) | 1 (`Docs/routes.md`) | 17 |

A regressao da comissao duplicada (P-01) tem teste dedicado: `tests/integration/order_test.go::TestOrderUpdateDoesNotDuplicateCommission`. Esse teste falha contra a implementacao C# e passa contra a implementacao Go - uma prova executavel da correcao.

## 5. Aspectos onde o Go nao traz vantagem

- **Maturidade do ecossistema corporativo**: ASP.NET Core tem suporte oficial da Microsoft, ferramentas de profiling, observability e identity-management bem integradas. Go exige composicao de bibliotecas externas para o mesmo resultado.
- **Geracao de documentacao OpenAPI**: o original usa Swashbuckle, que gera Swagger UI automaticamente a partir de atributos. Em Go, equivalente exigiria adicionar `swaggo/swag` ou similar - nao foi feito neste port.
- **Familiaridade da equipe**: depende do time. Times habituados ao ecossistema .NET podem ter custo inicial de aprendizado.

## 6. Avaliacao final de viabilidade

A migracao para Go e viavel e tecnicamente justificada para este projeto. Os ganhos em manutenibilidade (suite de testes), seguranca (bcrypt, JWT secret validado, omissao do password em resposta), integridade de dados (FKs e constraints) e operacao (imagem pequena, boot rapido) sao mensuraveis.

O custo da migracao - reescrita controlada, ~14 commits, ~3500 linhas de codigo Go novas - e proporcional ao escopo da CarShopApi (CRUD sobre 6 tabelas + auth). Para sistemas significativamente maiores, esse custo cresceria de forma mais que linear se a estrategia de "rewrite" fosse mantida; alternativas como evolucao incremental no proprio C# (corrigindo P-01 a P-13 sem trocar de linguagem) seriam mais defensaveis.

Para o caso especifico estudado, a migracao tambem serviu de instrumento didatico: cada problema identificado teve uma contraparte estrutural na nova versao, o que torna a comparacao educativa para a disciplina de Manutencao de Software.
