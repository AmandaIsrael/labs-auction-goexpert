# labs-auction-goexpert

Sistema de leilões em Go (Gin + MongoDB) com **fechamento automático de leilão** via
goroutine. Ao criar um leilão, uma rotina em background é agendada e, ao expirar a duração
configurada, o status do leilão é atualizado automaticamente para `Completed` no banco —
sem intervenção manual.

## Como rodar (Docker / Docker Compose)

Pré-requisito: Docker e Docker Compose instalados.

```bash
# sobe a API (porta 8080) + MongoDB (porta 27017)
docker compose up -d --build

# logs da aplicação
docker compose logs -f app

# derruba tudo
docker compose down            # mantém os dados
docker compose down -v         # apaga também o volume do Mongo
```

A API fica disponível em `http://localhost:8080`.

### Endpoints principais

| Método | Rota                          | Descrição                          |
|--------|-------------------------------|------------------------------------|
| POST   | `/auction`                    | Cria um leilão                     |
| GET    | `/auction?status=0`           | Lista leilões (0=Active, 1=Completed) |
| GET    | `/auction/:auctionId`         | Busca leilão por id                |
| GET    | `/auction/winner/:auctionId`  | Lance vencedor do leilão           |
| POST   | `/bid`                        | Cria um lance                      |
| GET    | `/bid/:auctionId`             | Lista lances de um leilão          |
| GET    | `/user/:userId`               | Busca usuário por id               |

Exemplo de criação de leilão (`condition`: 1=New, 2=Used, 3=Refurbished):

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{"product_name":"Notebook Dell","category":"informatica","description":"Notebook Dell i7 16GB","condition":2}'
```

## Variáveis de ambiente

Configuradas em `cmd/auction/.env` (carregadas pelo Docker Compose).

| Variável                | Exemplo                | Descrição                                                                 |
|-------------------------|------------------------|---------------------------------------------------------------------------|
| **`AUCTION_DURATION`**  | `20m`                  | **Duração do leilão.** Após esse tempo, o fechamento automático marca o leilão como `Completed`. |
| `AUCTION_INTERVAL`      | `20m`                  | Janela usada pela validação de lances para rejeitar lances em leilões expirados. |
| `BATCH_INSERT_INTERVAL` | `20s`                  | Intervalo de flush do lote de lances.                                     |
| `MAX_BATCH_SIZE`        | `4`                    | Tamanho do lote de lances antes do flush imediato.                        |
| `MONGODB_URL`           | `mongodb://...`        | String de conexão do MongoDB.                                             |
| `MONGODB_DB`            | `auctions`             | Nome do banco.                                                            |

### Formato e configuração do tempo

`AUCTION_DURATION` e `AUCTION_INTERVAL` usam o formato de `time.Duration` do Go:
`300ms`, `30s`, `20m`, `1h`, `1h30m`. Se o valor for inválido/ausente, o padrão é `20m`.

> **Importante:** mantenha **`AUCTION_DURATION` e `AUCTION_INTERVAL` com o mesmo valor**.
> O fechamento automático usa `AUCTION_DURATION`; a validação que impede lances em leilões
> fechados (`internal/infra/database/bid/create_bid.go`) usa `AUCTION_INTERVAL`. Valores
> divergentes fazem a janela de aceite de lances não coincidir com o momento do fechamento.

## Onde fica o fechamento automático

`internal/infra/database/auction/create_auction.go`:

- `CreateAuction` insere o leilão e dispara `go ar.scheduleAuctionClose(id)` (fire-and-forget,
  **não bloqueia** a requisição).
- `scheduleAuctionClose` usa um `time.NewTimer(AUCTION_DURATION)`, aguarda o canal `timer.C` e,
  ao expirar, chama `UpdateAuctionStatus` (em `update_auction.go`) com um **contexto novo** para setar `status = Completed`.

## Testes

O teste automatizado comprova o fechamento (cria leilão → aguarda a duração → confirma
`Completed`). Ele sobe um MongoDB efêmero via **testcontainers-go**, então **requer o Docker
ativo**:

```bash
go test ./internal/infra/database/auction/... -v
```

## Créditos

Projeto baseado no lab original da Full Cycle:
https://github.com/devfullcycle/labs-auction-goexpert
