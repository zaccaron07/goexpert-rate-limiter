
# GoExpert Rate Limiter

Projeto desenvolvido em Go para limitar requisições (rate limiting) utilizando Redis. Suporta limitação por IP e por token (API_KEY).

## Funcionalidades

- Limite de requisições por IP e por token (API_KEY)
- Bloqueio temporário após exceder o limite
- Cabeçalhos de resposta informando o status do rate limit
- Configuração flexível via variáveis de ambiente

## Como rodar localmente com Docker

```bash
docker-compose up --build
```

A API estará em `http://localhost:8080`.

## Variáveis de ambiente principais

| Variável | Valor padrão | Descrição |
|----------|--------------|-----------|
| `REDIS_ADDR` | `localhost:6379` | Host e porta do Redis |
| `IP_REQUESTS_PER_SECOND` | `10` | Máx. requisições por IP/segundo |
| `TOKEN_REQUESTS_PER_SECOND` | `10` | Máx. requisições por token/segundo |
| `IP_BLOCK_DURATION_SECONDS` | `5` | Duração do bloqueio para IP |
| `TOKEN_BLOCK_DURATION_SECONDS` | `5` | Duração do bloqueio para token |
| `SERVER_PORT` | `8080` | Porta do servidor HTTP |

## Endpoint exemplo

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET`  | `/`  | Rota protegida pelo limiter |

Requisição permitida:
```
HTTP/1.1 200 OK
X-Ratelimit-Remaining: <restante>
X-Ratelimit-Reset: <timestamp>
```

Quando o limite é excedido:
```
HTTP/1.1 429 Too Many Requests
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "block_until": "<timestamp>"
}
```

## Teste de carga rápido (k6)

```bash
# Por IP
k6 run tests/k6/rate_limit_test.js
# Por token
k6 run tests/k6/token_rate_limit_test.js
```

## Testes unitários

```bash
go test ./...
```
