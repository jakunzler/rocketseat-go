# Segurança

A API assume que quem chama pode ser um browser, um app ou outro serviço. O desenho abaixo é o padrão recomendado para esse uso.

## Senha

Senhas usam Argon2id com os parâmetros mínimos da OWASP: 19 MiB, 2 iterações e paralelismo 1. O hash guardado segue o formato PHC. A verificação compara o resultado em tempo constante e recusa parâmetros acima de um teto, para um hash adulterado não virar um vetor de negação de serviço.

O login de um e-mail inexistente ainda executa uma verificação contra um hash fictício, gerado na subida do processo. A resposta é sempre `invalid email or password`.

## Sessão

O access token é um JWT HS256 de curta duração (15 minutos por padrão). O segredo precisa ter pelo menos 32 bytes. O parser só aceita HS256 e o issuer configurado.

O refresh token é um valor aleatório de 32 bytes. O banco guarda só o SHA-256. Cada refresh revoga o token anterior e emite outro, na mesma transação. Reapresentar um token já revogado invalida todos os refresh tokens daquele usuário.

`POST /api/v1/auth/logout` revoga um refresh token. `POST /api/v1/auth/logout-all` exige o access token e revoga a família inteira.

## HTTP

- Cabeçalhos `nosniff`, `DENY` para frame, CSP sem recursos, `Referrer-Policy` e `Cache-Control: no-store`
- HSTS só quando `API_ENV=production`
- CORS limitado às origens de `API_CORS_ORIGINS`; `*` é recusado em produção
- Corpo JSON de no máximo 1 MiB, com campos desconhecidos rejeitados
- Limite global por IP e um limite mais baixo nas rotas `/api/v1/auth`
- `API_TRUST_PROXY` começa desligado. Só ligue atrás de um proxy que sobrescreve `X-Forwarded-For`
- Timeouts de leitura, escrita e de request context
- Recover de panic e log estruturado sem corpo nem header `Authorization`
- `/metrics` fica no mesmo processo. Publique essa rota só na rede interna

## Dados

Consultas usam parâmetros. O schema reforça tamanho, e-mail em minúsculas, prioridade e unicidade. Uma tarefa de outro usuário responde `404`, o mesmo código de um id inexistente.

`API_DATABASE_URL` em produção precisa de `sslmode=require`, `verify-ca` ou `verify-full`.

## O que ainda é decisão do projeto

Traduzir as mensagens, trocar o domínio de tarefas e proteger `/metrics` com rede ou um proxy. O template não coloca o refresh token em cookie: o corpo JSON funciona para mobile e para outros serviços. Um browser pode guardar o refresh token em memória do cliente e o access token só pelo tempo do `expires_at`.
