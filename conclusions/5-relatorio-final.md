# Relatorio final do projeto

Trabalho da disciplina de Manutencao de Software. Equipe: Andre Leoncio, Antonio Bordignon, Diego Contini e Matheus dos Santos. Data de referencia: 26/04/2026.

Este documento sintetiza o resultado das atividades previstas no projeto basico ("Analise e Evolucao de uma API em C# com Proposta de Migracao para Go") e referencia os entregaveis produzidos.

## 1. Sumario executivo

Foi analisada uma API REST em C# (CarShopApi) que atende uma aplicacao movel de gerenciamento de garagem de automoveis. A analise identificou 14 problemas de manutencao, sendo 5 de criticidade Alta. Como proposta de evolucao, foi implementada uma reescrita completa em Go (`carshop-api-golang/`) preservando o contrato HTTP original. A reescrita endereca todos os problemas identificados, inclui uma suite de 19 testes automatizados (antes inexistente) e reduz a imagem Docker em aproximadamente 92%.

## 2. Entregaveis

Os cinco entregaveis previstos na secao 11 do projeto basico foram produzidos como arquivos markdown nesta pasta:

| # | Documento | Arquivo |
|---|---|---|
| 1 | Documento de analise da API | [`1-analise-da-api.md`](1-analise-da-api.md) |
| 2 | Lista de problemas de manutencao identificados | [`2-problemas-identificados.md`](2-problemas-identificados.md) |
| 3 | Propostas de melhoria estrutural | [`3-propostas-melhoria.md`](3-propostas-melhoria.md) |
| 4 | Avaliacao da migracao para Go | [`4-avaliacao-migracao-go.md`](4-avaliacao-migracao-go.md) |
| 5 | Relatorio final do projeto | este documento |

Entregaveis adicionais produzidos no codigo:

- `../carshop-api-golang/` - implementacao Go completa, executavel.
- `../carshop-api-golang/bug-fixes-report/` - sete arquivos markdown, um por bug corrigido, com referencia ao codigo C# original e ao teste que verifica a correcao.
- `../carshop-api-golang/docs/routes.md` - contrato HTTP do servico Go.
- `../carshop-api-golang/claude-memory/` - documentacao tecnica auxiliar.

## 3. Status dos objetivos especificos

Mapeamento direto com a secao 4 do projeto basico:

| # | Objetivo | Documento que cumpre | Status |
|---|---|---|---|
| 1 | Mapear as funcionalidades principais da API | 1 (secoes 2 e 3) | Atendido |
| 2 | Identificar dificuldades de manutencao e organizacao do codigo | 2 (P-08 a P-12, P-14) | Atendido |
| 3 | Analisar limitacoes relacionadas a escalabilidade | 2 (P-02, P-03) | Atendido |
| 4 | Avaliar o impacto das limitacoes no uso e evolucao do sistema | 2 (coluna Impacto) + 4 (secao 4) | Atendido |
| 5 | Propor melhorias estruturais no sistema atual | 3 | Atendido |
| 6 | Analisar a viabilidade da migracao para Go | 4 | Atendido |
| 7 | Documentar os problemas encontrados e as solucoes propostas | 2 + `bug-fixes-report/` | Atendido |

## 4. Sintese das mudancas

### 4.1. Problemas resolvidos por criticidade

| Criticidade | Quantidade na CarShopApi | Resolvidos na versao Go |
|---|---|---|
| Alta | 5 | 5 (P-01, P-04, P-05, P-09, P-13) |
| Media | 6 | 6 (P-02, P-03, P-06, P-07, P-08, P-11) |
| Baixa | 3 | 3 (P-10, P-12, P-14) |

### 4.2. Indicadores antes/depois

| Metrica | Antes (C#) | Depois (Go) |
|---|---|---|
| Testes automatizados | 0 | 19 |
| Schema constraints (UNIQUE + FK) | 2 | 7 |
| Round trips para criar pedido com 5 itens | ~8 | 3 |
| Round trips para atualizar carro com 5 imagens | ~12 | 3 |
| Imagem Docker final | ~250 MB | ~20 MB |
| Arquivos `.md` de documentacao versionada | 1 | 17 |

### 4.3. Linhas de codigo

| Categoria | C# (linhas) | Go (linhas) |
|---|---|---|
| Codigo de aplicacao | ~1100 | ~2200 |
| Codigo gerado por ferramentas | ~250 (EF migrations) | ~900 (sqlc) |
| Codigo de teste | 0 | ~700 |
| Migrations SQL | 0 | ~50 |

A versao Go tem mais codigo total porque inclui suite de testes, codigo gerado pelo sqlc (que substitui o que EF Core faria em runtime) e documentacao. O codigo de aplicacao em si dobrou, principalmente por causa de:

- Camada `domain/` que nao existia.
- DTOs explicitos para request e response (com funcoes de conversao).
- Manuseio explicito de erros em vez de `try/catch` generico.

## 5. Riscos previstos vs riscos enfrentados

A secao 10 do projeto basico listou quatro riscos. Status:

| Risco | Estrategia prevista | Resultado |
|---|---|---|
| Dificuldade de entendimento da API | Divisao da analise por modulos | A organizacao por agregado (User, Car, Order, Comission) funcionou; cada documento e cada teste tem escopo definido. |
| Complexidade do codigo existente | Foco nas funcionalidades principais | A CarShopApi e relativamente pequena (~1100 LoC); a complexidade real estava nos problemas estruturais, nao no volume. |
| Tempo insuficiente | Definicao de escopo reduzido | Escopo foi mantido conforme planejado. |
| Falta de documentacao previa | Analise baseada no comportamento da API | O arquivo `CarShopApi/src/Docs/routes.md` existia e foi suficiente como base; o codigo confirmou o contrato. |

## 6. Licoes aprendidas

- **Documentacao versionada ao lado do codigo e barata e gera retorno alto**. O simples ato de manter `routes.md` no repositorio do C# foi o que permitiu a reescrita preservar o contrato sem ambiguidade.
- **Bugs de schema (P-05, P-07) sao mais caros que bugs de codigo**. Eles persistem ate o restart do banco e sao invisiveis no codigo. Constraints declaradas no SQL sao a forma mais barata de evitar essa classe.
- **Suite de testes de integracao e a infraestrutura mais valiosa de qualidade**. Testes unitarios cobrem regras de dominio; testes de integracao com banco real e router real cobrem o contrato. A diferenca entre 0 e 13 testes de integracao e mais relevante que qualquer otimizacao individual.
- **Reescritas funcionam quando o contrato externo e respeitado**. Os 14 commits incrementais, cada um deixando o sistema funcional, eliminaram a necessidade de manter dois sistemas em paralelo durante a transicao.

## 7. Limitacoes do trabalho

- A comparacao de performance e qualitativa, baseada em contagem de round trips e tamanho de imagem. Nao foram conduzidas medicoes de latencia ponta a ponta sob carga.
- A suite de testes em Go cobre o contrato HTTP e as regras de dominio, mas nao inclui testes de carga, testes de fuzzing nem testes de seguranca automatizados.
- A versao Go nao implementa documentacao OpenAPI / Swagger UI automatica, presente no original.
- A migracao foi avaliada para uma API pequena. Conclusoes sobre custo/beneficio nao se transferem diretamente para sistemas de maior escala.

## 8. Conclusao

A analise da CarShopApi mostrou um sistema funcional porem com debitos tecnicos significativos: 5 problemas criticos relacionados a seguranca, integridade de dados e ausencia de testes, alem de 9 problemas de criticidade Media/Baixa.

A migracao para Go, executada como reescrita controlada com preservacao do contrato HTTP, endereçou todos os problemas identificados e produziu um sistema mensuravelmente melhor em manutenibilidade, seguranca e desempenho.

Para a disciplina de Manutencao de Software, o exercicio cumpre seu papel didatico: cada item da analise tem correspondencia direta com uma decisao estrutural na versao evoluida, o que torna o codigo final um exemplo executavel das tecnicas estudadas.
