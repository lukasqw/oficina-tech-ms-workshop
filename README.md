# oficina-tech-ms-workshop

**MS3 — Workshop Service (Saga Participant)**

Microsserviço responsável pelo catálogo de serviços da oficina e pelo controle de estoque de produtos. Atua como **participante do Saga Pattern** orquestrado pelo MS2 (`ms-order-service`) — executa operações de estoque sob comando e publica o resultado de volta ao orquestrador.

---

## Arquitetura Interna

```
cmd/api/main.go
    │
    ├── internal/modules/
    │   ├── inventory/          ← produtos, estoque e operações do saga
    │   └── service_catalog/    ← catálogo de serviços da oficina
    │
    ├── internal/messaging/
    │   ├── consumers/          ← consome order-inventory-op-requested (SQS)
    │   └── publishers/         ← publica order-inventory-op-succeeded/failed (SQS/SNS)
    │
    ├── internal/shared/
    │   ├── dto/                ← DTOs compartilhados
    │   ├── infra/database/     ← conexão PostgreSQL (GORM)
    │   ├── infra/http/         ← middleware JWT e RBAC
    │   ├── infra/sqs_sns/      ← cliente SQS/SNS
    │   └── infra/observability/← OpenTelemetry
    │
    ├── migrations/             ← SQL migrations
    └── infra/terraform/        ← IaC do serviço
```

### Camadas por módulo

```
handler (HTTP) → usecase (negócio) → repository (interface) → infra/persistence (GORM)
                      ↓
                   entity (domínio puro)

messaging/consumers → usecase process_saga_operation → infra/persistence → messaging/publishers
```

---

## Módulos Internos

### inventory

Gestão de produtos e seus inventários, incluindo as operações determinadas pelo Saga.

**Operações de estoque (use cases):**

| Use Case | Operação Saga | Descrição |
|----------|--------------|-----------|
| `reserve_stock` | `RESERVE` | Move available → reserved (parcial: cria pending) |
| `reserved_decrease_stock` | `RESERVED_DECREASE` | Consome reserva (falha se insuficiente) |
| `cancel_reserved_stock` | `CANCEL_RESERVED` | Libera reserva/pending → available |
| `cancel_confirmed_stock` | `CANCEL_CONFIRMED` | Devolve estoque confirmado ao available |
| `increase_stock` | — | Entrada manual de mercadoria |
| `manual_decrease_stock` | — | Saída manual de estoque |
| `process_saga_operation` | — | Entry point do consumer: despacha para o use case correto |

**Lógica dos 3 estados do inventário:**

```
Estado: (available=A, reserved=R, pending=P)

RESERVE(q):
  q ≤ A  → (A-q, R+q, P)           reserva total
  q > A  → (0, R+A, P+(q-A))       reserva parcial + backorder

RESERVED_DECREASE(q):
  R ≥ q  → (A, R-q, P)             consome reserva
  R < q  → ERRO ErrInsufficientReserved

CANCEL_RESERVED(q):
  P ≥ q  → (A, R, P-q)             cancela backorder
  P < q  → (A+(q-P), R-(q-P), 0)  cancela backorder + devolve reserva

CANCEL_CONFIRMED(q):
  P > 0, P ≥ q → (A, R, P-q)      abate pending
  P > 0, P < q → (A+(q-P), R+P, 0)
  P = 0        → (A+q, R, P)       devolve ao disponível

INCREASE(q):
  P ≥ q → (A, R+q, P-q)           abate backorder
  P < q → (A+(q-P), R+P, 0)       abate parcial + adiciona disponível
```

**Idempotência via `saga_operations`:**
Antes de processar qualquer operação, verifica se `saga_id + operation` já existe na tabela. Se sim, republica o resultado sem reprocessar — garante segurança com at-least-once delivery do SQS.

**CRUD de produtos:** create, get, get_all, update, delete product.

### service_catalog

Catálogo dos serviços oferecidos pela oficina (ex: "Troca de óleo", "Revisão de freios").

| Use Case | Descrição |
|----------|-----------|
| `create_service` | Cria serviço no catálogo |
| `get_service` / `get_all` | Consulta serviço(s) |
| `update_service` | Atualiza dados do serviço |
| `delete_service` | Remove serviço |

Entidade de domínio: `Service` (id, name, description, price, estimated_duration_minutes, category, is_active)

---

## Banco de Dados

**PostgreSQL** — `db_ms3` (porta local `5435`)

```
services (
  id UUID PK, name, description, price,
  estimated_duration_minutes, category, is_active,
  created_at, updated_at, deleted_at
)

products (
  id UUID PK, code, name, description,
  price, product_type,
  created_at, updated_at, deleted_at
)

inventories (
  id UUID PK, product_id FK → products,
  available_quantity,   -- pronto para uso
  reserved_quantity,    -- alocado em OS, não consumido
  pending_quantity,     -- prometido em OS, aguardando entrada
  min_quantity,         -- threshold de alerta de estoque baixo
  created_at, updated_at
)

saga_operations (
  id UUID PK, saga_id, order_id,
  operation,   -- RESERVE | RESERVED_DECREASE | CANCEL_RESERVED | CANCEL_CONFIRMED
  status,      -- PROCESSING | COMPLETED | FAILED
  processed_at
)
-- saga_operations garante idempotência: saga_id + operation é chave única
```

Schema gerenciado via migrations SQL em `migrations/`.

---

## Endpoints HTTP

**Porta**: `8083`

```
POST   /services                        Criar serviço
GET    /services                        Listar serviços
GET    /services/{id}                   Buscar serviço ← consultado pelo MS2 (snapshot)
PUT    /services/{id}                   Atualizar serviço
DELETE /services/{id}                   Remover serviço
PATCH  /services/{id}/toggle            Ativar/desativar serviço

POST   /products                        Criar produto
GET    /products                        Listar produtos
GET    /products/{id}                   Buscar produto ← consultado pelo MS2 (snapshot)
PUT    /products/{id}                   Atualizar produto
DELETE /products/{id}                   Remover produto

GET    /inventory                       Listar todos os inventários
GET    /inventory/{product_id}          Inventário de um produto
POST   /inventory                       Criar inventário para produto
POST   /inventory/{product_id}/increase Entrada manual de estoque
POST   /inventory/{product_id}/decrease Saída manual de estoque
```

---

## Eventos

### Consumidos (SQS)

```json
// Fila: order-inventory-op-requested
// Publicado pelo MS2 a cada saga step
{
  "event": "OrderInventoryOpRequested",
  "saga_id": "uuid",
  "order_id": "uuid",
  "operation": "RESERVE",   // RESERVE | RESERVED_DECREASE | CANCEL_RESERVED | CANCEL_CONFIRMED
  "items": [
    { "product_id": "uuid", "quantity": 2 }
  ],
  "occurred_at": "RFC3339"
}
```

### Publicados (SQS/SNS)

```json
// Tópico: order-inventory-op-succeeded
{
  "event": "OrderInventoryOpSucceeded",
  "saga_id": "uuid",
  "order_id": "uuid",
  "operation": "RESERVE",
  "occurred_at": "RFC3339"
}

// Tópico: order-inventory-op-failed
{
  "event": "OrderInventoryOpFailed",
  "saga_id": "uuid",
  "order_id": "uuid",
  "operation": "RESERVE",
  "reason": "Produto 'Filtro de óleo' sem estoque: disponível=1, solicitado=2",
  "occurred_at": "RFC3339"
}

// Tópico: inventory-low-alert  (side effect não bloqueante)
{
  "event": "InventoryLowAlert",
  "product_id": "uuid",
  "product_name": "string",
  "available_quantity": 1,
  "reserved_quantity": 3,
  "pending_quantity": 2,
  "min_quantity": 5,
  "occurred_at": "RFC3339"
}
```

---

## Variáveis de Ambiente

```env
SERVER_PORT=8083

DB_HOST=localhost
DB_PORT=5435
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=db_ms3

JWT_SECRET_KEY=               # compartilhado com todos os serviços

AWS_REGION=us-east-1
AWS_ENDPOINT=http://localhost:4566    # LocalStack em dev
SQS_INVENTORY_OP_REQUESTED_URL=      # fila a consumir (MS2 → MS3)
SNS_INVENTORY_OP_RESULT_ARN=         # tópico a publicar (MS3 → MS2)

OTEL_EXPORTER_OTLP_ENDPOINT=         # opcional — observabilidade
```

---

## Como Rodar Localmente

```bash
# Sobe PostgreSQL + LocalStack (SQS + SNS)
docker-compose up -d

# Aplica migrations
# (executado automaticamente na inicialização via scripts/)

# Instala dependências
go mod download

# Roda o serviço
go run cmd/api/main.go
```

A API fica disponível em `http://localhost:8083`.

---

## Testes

```bash
# Unitários — foco nas operações de estoque e idempotência do saga
go test ./internal/modules/inventory/application/usecases/...

# Integração — handlers e consumer
go test ./internal/modules/.../infra/http/handlers/...

# Cobertura mínima: 80%
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Cenários críticos de cobertura:
- Todas as operações de estoque (RESERVE, RESERVED_DECREASE, CANCEL_RESERVED, CANCEL_CONFIRMED)
- Idempotência: operação duplicada republica resultado sem reprocessar
- Reserva parcial (backorder) e recuperação

---

## Infraestrutura

```
k8s/
├── namespace.yaml
├── deployment.yaml         2 réplicas, RollingUpdate, probes
├── service.yaml            ClusterIP → NodePort 30083
├── hpa.yaml
├── configmap.yaml.example
└── secret.yaml.example

infra/terraform/            IaC do serviço
migrations/                 SQL migrations (schema do banco)
```

Pipeline CI/CD: `.github/workflows/` — `ci.yml` → `deploy.yml` → `release.yml` | `rollback.yml`
