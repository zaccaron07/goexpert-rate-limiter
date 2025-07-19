
# GoExpert Rate Limiter

Projeto desenvolvido em Go para limitar requisições (rate limiting) utilizando Redis. Suporta limitação por IP e por token (API_KEY).

## Funcionalidades

- Limite de requisições por IP e por token (API_KEY)
- Bloqueio temporário após exceder o limite
- Cabeçalhos de resposta informando o status do rate limit
- Configuração flexível via variáveis de ambiente

## Endpoints

- `GET /` — Endpoint de exemplo protegido pelo rate limiter.

## Respostas e Cabeçalhos

Quando permitido:
```
HTTP 200 OK
X-Ratelimit-Remaining: <quantidade restante>
X-Ratelimit-Reset: <timestamp de reset>
```

Quando bloqueado:
```
HTTP 429 Too Many Requests
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "block_until": "<timestamp>"
}
```

## Configuração

- Edite as variáveis no arquivo `env.example` e renomeie para `.env` na raiz do projeto.
- Para testar limite por token, envie o header `API_KEY` na requisição.

## Como rodar localmente com Docker

1. Suba os containers (aplicação + Redis):
   ```sh
   docker-compose up --build
   ```
2. Acesse a API em http://localhost:8080

## Como rodar os testes de carga com k6

Você pode rodar os testes de carga localmente ou via Docker Compose:

- Instale o [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/).

- Para o teste de limite por IP:
  ```sh
  k6 run tests/k6/rate_limit_test.js
  ```

- Para o teste de limite por token:
  ```sh
  k6 run tests/k6/token_rate_limit_test.js
  ```

---
