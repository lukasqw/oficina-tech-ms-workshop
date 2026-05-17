# oficina-tech-ms-workshop

**MS3 — Workshop Service (Participante do Saga Pattern)**

Microsserviço responsável pelo catálogo de serviços da oficina automotiva e pelo controle de estoque de produtos. Atua como participante do Saga Pattern orquestrado pelo MS2 (`ms-order-service`): recebe comandos de operação de estoque via SQS, executa as transições de estado e publica o resultado de volta ao orquestrador.

---

## Responsabilidades

- Manter o catálogo de serviços oferecidos pela oficina (criação, consulta, atualização, remoção)
- Manter o cadastro de produtos e seus inventários
- Controlar os três estados de quantidade de estoque: `available`, `reserved` e `pending`
- Executar operações de estoque determinadas pelo Saga: `RESERVE`, `RESERVED_DECREASE`, `CANCEL_RESERVED`, `CANCEL_CONFIRMED`
- Garantir idempotência das operações saga via tabela `saga_operations`
- Emitir alertas de estoque baixo (`inventory-low-alert`) como efeito colateral não bloqueante

---

## Arquitetura Interna

```
cmd/api/main.go
    │
    ├── internal/modules/
    │   ├── inventory/                  produtos, inventários e operações do saga
    │   │   ├── domain/                 entidades puras e interfaces de repositório
    │   │   │   ├── inventory/
    │   │   │   ├── product/
    │   │   │   └── saga_operation/
    │   │   ├── application/usecases/   lógica de negócio
    │   │   └── infra/
    │   │       ├── http/handlers/      handlers HTTP
    │   │       ├── http/dto/           DTOs de request/response
    │   │       ├── http/routes.go      registro de rotas
    │   │       └── persistence/        modelos GORM e implementações de repositório
    │   │
    │   └── service_catalog/            catálogo de serviços da oficina
    │       ├── domain/service/
    │       ├── application/usecases/
    │       └── infra/
    │           ├── http/handlers/
    │           ├── http/dto/
    │           ├── http/routes.go
    │           └── persistence/
    │
    ├── internal/messaging/
    │   ├── consumers/                  consome fila order-inventory-op-requested
    │   └── publishers/                 publica em order-inventory-op-succeeded e order-inventory-op-failed
    │
    └── internal/shared/
        ├── dto/                        DTOs compartilhados com outros serviços
        ├── infra/auth/                 validação JWT local (HS256)
        ├── infra/database/             conexão PostgreSQL via GORM
        ├── infra/http/middleware/      middlewares de autenticação, RBAC e observabilidade
        ├── infra/http/routes/          setup central de rotas e health check
        ├── infra/awsconfig/            configuração do SDK AWS v2
        ├── infra/sqs/                  cliente SQS
        ├── infra/observability/        OpenTelemetry (traces, métricas, logs)
        └── utils/                      respostas HTTP padronizadas, error codes
```

### Fluxo de camadas

```
handler (HTTP) → usecase → repository (interface) → persistence (GORM) → PostgreSQL

messaging/consumer → usecase process_saga_operation → persistence → messaging/publisher
```

---

## Endpoints HTTP

**Porta padrão:** `8083`

### Serviços (`/services`)

| Método | Path | Roles | Descrição |
|--------|------|-------|-----------|
| `POST` | `/services` | MANAGER, ADMIN | Criar serviço no catálogo |
| `GET` | `/services` | USER, MANAGER, ADMIN | Listar todos os serviços |
| `GET` | `/services/{id}` | USER, MANAGER, ADMIN | Buscar serviço por ID (consultado pelo MS2 em snapshot) |
| `PUT` | `/services/{id}` | MANAGER, ADMIN | Atualizar dados do serviço |
| `DELETE` | `/services/{id}` | MANAGER, ADMIN | Remover serviço |

### Produtos (`/products`)

| Método | Path | Roles | Descrição |
|--------|------|-------|-----------|
| `POST` | `/products` | USER, MANAGER, ADMIN | Criar produto |
| `GET` | `/products` | USER, MANAGER, ADMIN | Listar todos os produtos |
| `GET` | `/products/{id}` | USER, MANAGER, ADMIN | Buscar produto por ID (consultado pelo MS2 em snapshot) |
| `PUT` | `/products/{id}` | USER, MANAGER, ADMIN | Atualizar produto |
| `DELETE` | `/products/{id}` | USER, MANAGER, ADMIN | Remover produto |

### Inventário (`/products/{product_id}/inventory`)

| Método | Path | Roles | Descrição |
|--------|------|-------|-----------|
| `POST` | `/products/{product_id}/inventory` | MANAGER, ADMIN | Criar inventário para produto |
| `GET` | `/products/{product_id}/inventory` | USER, MANAGER, ADMIN | Consultar inventário de um produto |
| `DELETE` | `/products/{product_id}/inventory` | MANAGER, ADMIN | Remover inventário |
| `POST` | `/products/{product_id}/inventory/increase` | MANAGER, ADMIN | Entrada manual de estoque |
| `POST` | `/products/{product_id}/inventory/manual-decrease` | MANAGER, ADMIN | Saída manual de estoque |
| `POST` | `/products/{product_id}/inventory/available-decrease` | MANAGER, ADMIN | Diminuição direta do disponível |
| `POST` | `/products/{product_id}/inventory/reserve` | MANAGER, ADMIN | Reserva de estoque (saga RESERVE) |
| `POST` | `/products/{product_id}/inventory/reserved-decrease` | MANAGER, ADMIN | Consumo de reserva (saga RESERVED_DECREASE) |
| `POST` | `/products/{product_id}/inventory/cancel-reserved` | MANAGER, ADMIN | Cancelamento de reserva/backorder (saga CANCEL_RESERVED) |
| `POST` | `/products/{product_id}/inventory/cancel-confirmed` | MANAGER, ADMIN | Devolução de estoque confirmado (saga CANCEL_CONFIRMED) |

### Utilitários

| Método | Path | Auth | Descrição |
|--------|------|------|-----------|
| `GET` | `/health` | Nenhuma | Retorna status do serviço, banco de dados, versão e uptime |
| `GET` | `/swagger/index.html` | Nenhuma | Documentação OpenAPI (Swagger UI) |

---

## Saga Pattern — Participante

O MS3 é participante do saga orquestrado pelo MS2. Toda comunicação é assíncrona via SQS.

### Filas e tópicos

| Direção | Nome | Descrição |
|---------|------|-----------|
| Consumida | `order-inventory-op-requested` | Comandos de operação enviados pelo MS2 |
| Publicada | `order-inventory-op-succeeded` | Resultado positivo enviado ao MS2 |
| Publicada | `order-inventory-op-failed` | Resultado negativo enviado ao MS2 |
| Publicada (SNS) | `inventory-low-alert` | Alerta de estoque abaixo do mínimo (efeito colateral não bloqueante) |

### Operações suportadas

| Operação (`operation`) | Use Case | Comportamento |
|------------------------|----------|---------------|
| `RESERVE` | `reserve_stock` | Move `available → reserved`; se `available` insuficiente, cria backorder em `pending` |
| `RESERVED_DECREASE` | `reserved_decrease_stock` | Consome a reserva (`reserved - q`); falha com `ErrInsufficientReserved` se reserva menor que o pedido |
| `CANCEL_RESERVED` | `cancel_reserved_stock` | Cancela backorder e/ou reserva, devolvendo ao `available` |
| `CANCEL_CONFIRMED` | `cancel_confirmed_stock` | Devolve estoque já confirmado; abate `pending` antes de incrementar `available` |

### Idempotência

Antes de processar qualquer operação, o use case `process_saga_operation` verifica na tabela `saga_operations` se o par `(saga_id, operation)` já existe. Se sim, republica o resultado armazenado sem reprocessar — garante segurança com entrega at-least-once do SQS.

### Formato do evento consumido

```json
{
  "event": "OrderInventoryOpRequested",
  "saga_id": "uuid",
  "order_id": "uuid",
  "operation": "RESERVE",
  "items": [
    { "product_id": "uuid", "quantity": 2 }
  ],
  "occurred_at": "2026-05-17T10:00:00Z"
}
```

---

## Modelo de Estoque

Cada produto possui um registro na tabela `inventories` com três campos de quantidade:

| Campo | Significado |
|-------|-------------|
| `available_quantity` | Quantidade disponível para novos pedidos |
| `reserved_quantity` | Quantidade alocada em ordens de serviço abertas, ainda não consumida |
| `pending_quantity` | Quantidade prometida em ordens mas aguardando entrada física (backorder) |

### Transições de estado por operação

```
RESERVE(q):
  q <= available  →  available -= q, reserved += q
  q >  available  →  pending += (q - available), reserved += available, available = 0

RESERVED_DECREASE(q):
  reserved >= q   →  reserved -= q
  reserved <  q   →  ERRO: ErrInsufficientReserved

CANCEL_RESERVED(q):
  pending >= q    →  pending -= q
  pending <  q    →  available += (q - pending), reserved -= (q - pending), pending = 0

CANCEL_CONFIRMED(q):
  pending > 0, pending >= q  →  pending -= q
  pending > 0, pending <  q  →  available += (q - pending), reserved += pending, pending = 0  [nota: reserved retorna]
  pending = 0                →  available += q

INCREASE(q):
  pending >= q  →  reserved += q, pending -= q
  pending <  q  →  reserved += pending, available += (q - pending), pending = 0
```

---

## Banco de Dados

**PostgreSQL 16** — banco `db_ms3`, porta local `5435` (via docker-compose).

Schema gerenciado automaticamente por GORM (`AutoMigrate`) e pela migration SQL em `migrations/`.

### Tabelas

#### `services`
| Coluna | Tipo | Restrições |
|--------|------|-----------|
| `id` | UUID | PK, `gen_random_uuid()` |
| `name` | VARCHAR(100) | NOT NULL, UNIQUE |
| `description` | VARCHAR(500) | NOT NULL |
| `price` | INTEGER | NOT NULL (centavos) |
| `created_at` | TIMESTAMP | NOT NULL |
| `updated_at` | TIMESTAMP | NOT NULL |
| `deleted_at` | TIMESTAMP | índice (soft delete) |

#### `product_models`
| Coluna | Tipo | Restrições |
|--------|------|-----------|
| `id` | UUID | PK |
| `name` | VARCHAR(200) | NOT NULL |
| `description` | VARCHAR(1000) | |
| `price` | INTEGER | NOT NULL (centavos) |
| `product_type` | VARCHAR(20) | NOT NULL |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |
| `deleted_at` | TIMESTAMP | índice (soft delete) |

#### `inventories`
| Coluna | Tipo | Restrições |
|--------|------|-----------|
| `id` | UUID | PK |
| `product_id` | UUID | NOT NULL, UNIQUE (FK → product_models) |
| `available_quantity` | INTEGER | NOT NULL, default 0 |
| `reserved_quantity` | INTEGER | NOT NULL, default 0 |
| `pending_quantity` | INTEGER | NOT NULL, default 0 |
| `created_at` | TIMESTAMP | NOT NULL |
| `updated_at` | TIMESTAMP | NOT NULL |
| `deleted_at` | TIMESTAMP | índice (soft delete) |

#### `saga_operations`
| Coluna | Tipo | Restrições |
|--------|------|-----------|
| `id` | UUID | PK |
| `saga_id` | UUID | NOT NULL |
| `order_id` | UUID | NOT NULL |
| `operation` | VARCHAR(40) | NOT NULL |
| `status` | VARCHAR(40) | NOT NULL (`PROCESSING`, `COMPLETED`, `FAILED`) |
| `result_payload` | JSONB | |
| `processed_at` | TIMESTAMP | |
| — | — | UNIQUE (`saga_id`, `operation`) |

---

## Variáveis de Ambiente

```env
# Servidor
SERVER_PORT=8083

# Banco de Dados
DB_HOST=localhost
DB_PORT=5435
DB_USER=oficina
DB_PASSWORD=oficina
DB_NAME=db_ms3
DB_SSLMODE=disable

# JWT — mesmo segredo compartilhado com MS1 e MS2 (validação local, HS256)
JWT_SECRET_KEY=local-dev-secret

# AWS — LocalStack (desenvolvimento local)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_ENDPOINT_URL=http://localhost:4566

# AWS Academy (produção) — substituir pelas credenciais da sessão
# AWS_ACCESS_KEY_ID=<copiado do AWS Details>
# AWS_SECRET_ACCESS_KEY=<copiado do AWS Details>
# AWS_SESSION_TOKEN=<copiado do AWS Details>
# AWS_REGION=us-east-1
# (remover AWS_ENDPOINT_URL para apontar para a AWS real)

# SQS
ORDER_INVENTORY_OP_REQUESTED_QUEUE_URL=http://localhost:4566/000000000000/order-inventory-op-requested
ORDER_INVENTORY_OP_SUCCEEDED_QUEUE_URL=http://localhost:4566/000000000000/order-inventory-op-succeeded
ORDER_INVENTORY_OP_FAILED_QUEUE_URL=http://localhost:4566/000000000000/order-inventory-op-failed

# SNS
SNS_INVENTORY_LOW_ALERT_ARN=arn:aws:sns:us-east-1:000000000000:inventory-low-alert

# Observabilidade (opcional)
OTEL_EXPORTER_OTLP_ENDPOINT=
```

Copie `.env.example` para `.env` e preencha os valores antes de rodar localmente.

---

## Como Rodar Localmente

**Pré-requisitos:** Go 1.25, Docker, Docker Compose.

```bash
# 1. Copiar variáveis de ambiente
cp .env.example .env

# 2. Subir PostgreSQL e LocalStack (SQS + SNS)
docker-compose up -d

# 3. Instalar dependências
go mod download

# 4. Iniciar o serviço
go run cmd/api/main.go
```

A API fica disponível em `http://localhost:8083`.
A documentação Swagger fica disponível em `http://localhost:8083/swagger/index.html`.

Para rodar apenas as dependências (sem o serviço em container):

```bash
docker-compose up -d postgres localstack
go run cmd/api/main.go
```

---

## Testes

```bash
# Todos os testes com cobertura
go test ./... -v -coverprofile=coverage.out -covermode=atomic

# Relatório de cobertura em HTML
go tool cover -html=coverage.out

# Apenas usecases do módulo inventory
go test ./internal/modules/inventory/application/usecases/...

# Apenas handlers HTTP
go test ./internal/modules/inventory/infra/http/handlers/...
go test ./internal/modules/service_catalog/infra/http/handlers/...

# Vet e lint
go vet ./...
```

**Limiar de cobertura:** 60% (gate no CI). Cenários obrigatoriamente cobertos:

- Todas as quatro operações saga (RESERVE, RESERVED_DECREASE, CANCEL_RESERVED, CANCEL_CONFIRMED)
- Idempotência: operação duplicada republica resultado sem reprocessar
- Reserva parcial com backorder e recuperação via INCREASE
- Falha por reserva insuficiente (`ErrInsufficientReserved`)

---

## CI/CD

O pipeline é composto por quatro workflows independentes em `.github/workflows/`:

### `ci.yml` — Integração Contínua

Disparado em pull requests para `develop` e `main`, e em push para `develop`.

1. Checkout e setup do Go 1.25
2. `go mod download`
3. `go vet ./...`
4. `golangci-lint` (versão latest)
5. `go test ./... -coverprofile=coverage.out`
6. Gate de cobertura: falha se cobertura total abaixo de 60%

### `release.yml` — Criação de Release PR

Disparado em push para `develop` ou manualmente.

1. Verifica se já existe PR de release aberto (`release/*` → `main`)
2. Se existe: atualiza o PR com os novos commits de `develop`
3. Se não existe: calcula nova versão via Conventional Commits (`feat:` → minor, `feat!:` → major, demais → patch) e abre PR `release/vX.Y.Z` → `main`

### `deploy.yml` — Deploy em Produção

Disparado quando o PR de release é mergeado para `main`, ou manualmente via `workflow_dispatch`.

Estágios em sequência:

1. **Build & Test** — compila o binário e executa os testes; calcula a versão final
2. **Build & Push Docker** — constrói a imagem multi-stage e publica no GHCR (`ghcr.io/{repo}:{version}`)
3. **Deploy to Kubernetes** — conecta ao EKS via Terraform remote state (S3), aplica ConfigMap e Secret, aplica os manifestos K8s com `envsubst`, aguarda rollout e valida o health check em `/health`
4. **Finalize Release Tag** — cria a tag `vX.Y.Z` no repositório somente após o health check confirmar o deploy saudável
5. **GitHub Release** — publica o release no GitHub a partir da tag criada no passo anterior
6. **SonarQube** — análise de qualidade de código (paralela ao release)

### `rollback.yml` — Rollback Manual

Disparado manualmente com input `version` (ex: `v1.2.3`).

1. Valida o formato da versão
2. Faz checkout na tag alvo
3. Valida que a tag existe e que a imagem Docker correspondente existe no GHCR
4. Lê o estado atual da infraestrutura via Terraform remote state (S3)
5. Reaaplica os manifestos K8s da versão alvo com a imagem do GHCR
6. Aguarda rollout e valida health check
7. Não cria nova tag

---

## Deploy Kubernetes

Manifestos em `k8s/`:

| Arquivo | Recurso | Detalhes |
|---------|---------|---------|
| `namespace.yaml` | Namespace | `app-oficina-tech` |
| `deployment.yaml` | Deployment | 1 réplica inicial, RollingUpdate (`maxSurge=1`, `maxUnavailable=0`), `terminationGracePeriodSeconds=30` |
| `service.yaml` | Service | Tipo NodePort, porta `8083`, nodePort `30083` |
| `hpa.yaml` | HorizontalPodAutoscaler | mínimo 1, máximo 5 réplicas; escala com CPU > 70% |

### Recursos por container

| Recurso | Request | Limit |
|---------|---------|-------|
| Memória | 256Mi | 768Mi |
| CPU | 250m | 500m |

### Probes

| Probe | Path | Delay inicial | Período |
|-------|------|--------------|---------|
| liveness | `/health` | 60s | 10s |
| readiness | `/health` | 20s | 5s |
| startup | `/health` | — | 5s (máx 36 falhas = 3 min) |

Configurações de ambiente injetadas via ConfigMap e Secret referenciados no Deployment. A imagem é publicada no GHCR: `ghcr.io/{repositório}:{versão}`. A autenticação ao registry é feita via Secret `ghcr-secret`. Observabilidade via Datadog: o endpoint OTLP é configurado dinamicamente via `status.hostIP` do pod.
