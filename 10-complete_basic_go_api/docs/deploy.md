# Deploy

A imagem de produção é estática, roda como usuário não-root e não contém `.env`. Configuração e segredo entram pelo ambiente da plataforma. Migrações estão dentro do binário.

## Local

```bash
docker compose up --build
```

`/healthz` responde se o processo está vivo. `/readyz` consulta o Postgres. O healthcheck da imagem chama `/api healthcheck`, então a imagem distroless não precisa de shell nem de curl.

## GitHub Actions

Neste monorepo, o workflow que executa é `.github/workflows/complete-basic-go-api.yml`, na raiz. Ele formata, roda `go vet`, testes com `-race`, a suíte de integração contra Postgres 18 e `govulncheck`.

A pasta `.github/workflows` daqui dentro vale quando este diretório vira a raiz de um repositório próprio. `ci.yml` repete a verificação. `release.yml` publica a imagem no GHCR ao criar uma tag `vX.Y.Z`:

```text
ghcr.io/<owner>/<repo>:vX.Y.Z
```

## Migrações

`API_AUTO_MIGRATE=true` aplica o Goose na subida. Isso serve para uma instância só, inclusive o Compose local.

Com mais de uma réplica, rode a migração como passo de release e suba a aplicação com `API_AUTO_MIGRATE=false`. O processo em produção registra um aviso se a migração automática continuar ligada.

## Plataforma

Qualquer runtime que execute container e injete ambiente serve. O contrato operacional é:

- escutar `API_HTTP_ADDR` (padrão `:8080`)
- liveness: `GET /healthz`
- readiness: `GET /readyz`
- métricas: `GET /metrics`, de preferência numa rede privada
- segredo `API_JWT_SECRET` com 32 bytes ou mais, diferente do valor de `.env.example`
- `API_DATABASE_URL` com TLS
- `API_TRUST_PROXY=true` somente atrás de um load balancer confiável
- `API_ENV=production`

No Kubernetes, use `httpGet` em `/healthz` e `/readyz`. No Cloud Run ou no Fly.io, aponte o health check para `/healthz` e passe as variáveis pelo secret manager da plataforma. Não copie `.env` para a imagem: o exemplo antigo do curso fazia isso, e um segredo no layer da imagem continua recuperável.

## Build manual

```bash
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t go-api:1.0.0 .
```

`GET /api/v1/version` devolve esses valores.
