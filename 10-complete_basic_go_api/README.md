# API Go reutilizável

Base pronta para copiar em outros projetos. Ela junta o que as aulas construíram — pacotes, HTTP com Chi, validação, repositório, Postgres, concorrência, testes e deploy — com o que hoje é prática recomendada de segurança e operação.

O domínio de exemplo é usuários e tarefas. A borda HTTP, a autenticação, a configuração e o acesso ao banco ficam estáveis quando o domínio mudar.

## Stack

- Go 1.26, `slog`, `errgroup` e shutdown gracioso
- Chi para rotas `/api/v1`
- pgx e migrações embutidas com Goose
- Argon2id para senha e JWT de acesso com refresh token rotativo
- Prometheus em `/metrics`
- Imagem distroless, sem root, com healthcheck no próprio binário

## Como rodar

Sem banco, só para explorar a API:

```bash
cp .env.example .env
make dev
```

`make dev` sobe com armazenamento em memória. Os dados somem quando o processo para. Produção recusa esse modo.

Com Postgres:

```bash
cp .env.example .env
docker compose up --build
```

A API escuta em `http://localhost:8080`. No primeiro start do banco também é criado o database `api_test`, usado pelos testes de integração.

## Fluxo mínimo

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada Lovelace","email":"ada@example.com","password":"correct-horse"}'

curl -s http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer ACCESS_TOKEN"

curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{"title":"Estudar Go","description":"HTTP, SQL e testes","priority":2}'
```

## Endpoints

| Método | Caminho | Auth |
| --- | --- | --- |
| GET | `/healthz` | não |
| GET | `/readyz` | não |
| GET | `/metrics` | não |
| GET | `/api/v1/version` | não |
| POST | `/api/v1/auth/signup` | não |
| POST | `/api/v1/auth/login` | não |
| POST | `/api/v1/auth/refresh` | não |
| POST | `/api/v1/auth/logout` | não |
| POST | `/api/v1/auth/logout-all` | sim |
| GET | `/api/v1/me` | sim |
| GET | `/api/v1/tasks` | sim |
| POST | `/api/v1/tasks` | sim |
| GET | `/api/v1/tasks/{id}` | sim |
| PATCH | `/api/v1/tasks/{id}` | sim |
| DELETE | `/api/v1/tasks/{id}` | sim |

Respostas de sucesso usam `{"data": ...}`. Listagens incluem `meta.limit`, `meta.offset` e `meta.total`. Erros usam `{"error":{"code","message","request_id","fields"}}`. Os códigos (`validation_error`, `invalid_credentials`, `rate_limited`) são estáveis; as mensagens estão em inglês para o template servir em outros projetos.

O contrato está em [docs/openapi.yaml](docs/openapi.yaml).

## Testes

```bash
make test-race
make test-integration
```

`make test` cobre validação, Argon2id, JWT, serviços, handlers e o limitador, sem Postgres. A suíte de integração sobe o schema, grava usuário, tarefa e rotação de refresh token, e recusa um database cujo nome não contenha `test`.

## Estrutura

```text
cmd/api                         processo HTTP, sinais e healthcheck
internal/config                 ambiente e travas de produção
internal/platform               logger e http.Server
internal/httpapi                rotas, middleware, JSON
internal/auth                   Argon2id, JWT e refresh token
internal/validator              regras de entrada
internal/service                casos de uso
internal/store                  interfaces
internal/store/memory           adapter em memória
internal/store/postgres         pgx, Goose e SQL
internal/app                    composição das dependências
```

Para um domínio novo, troque `task` por outra entidade e mantenha `config`, `httpapi`, `auth` e `platform`. O module path atual é `github.com/jakunzler/rocketseat-go/10-complete_basic_go_api`; ao extrair a pasta, rode `go mod edit -module` com o caminho do novo repositório.

## O que veio das aulas

- Fundamentos: pacotes, erros, structs e testes de unidade
- Conceitos avançados: interfaces e um helper JSON genérico
- API REST: Chi, handlers e envelope JSON
- Banco: repositório, SQL parametrizado e migração
- HTTP: middleware, versão `/api/v1`, auth e shutdown
- Concorrência: `errgroup`, worker de limpeza e mutex no store em memória
- Testes: tabela, `httptest`, store falso e integração com Postgres
- Deploy: imagem multi-stage, Compose e GitHub Actions

Detalhes de segurança e de publicação: [docs/seguranca.md](docs/seguranca.md) e [docs/deploy.md](docs/deploy.md).

## Variáveis

O arquivo [`.env.example`](.env.example) lista todas. Em produção o processo recusa segredo de desenvolvimento, `API_STORE=memory`, `sslmode=disable` e CORS `*`.

Gere o segredo com:

```bash
openssl rand -base64 48
```
