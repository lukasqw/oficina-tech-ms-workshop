# MS3 — Cobertura

| Escopo | Cobertura | Meta | Resultado |
|---|---|---|---|
| `internal/modules/inventory/...` (product, inventory, saga_operation, todos os 7 use cases de estoque + ProcessSagaOperation) | **81.4 %** | ≥ 80 % | ✅ |
| Total `internal/...` (incluindo shared/database, observability, sqs) | 69.2 % | — | informativo |

Pacotes ainda em 0 %: wiring HTTP (`internal/shared/infra/http/routes`), `observability`, `sqs` (helpers de cliente AWS) — exercitados pela suíte BDD e2e.

## Regenerar

```sh
go test -count=1 -covermode=atomic -coverprofile=docs/coverage/coverage.out ./internal/...
go tool cover -html=docs/coverage/coverage.out -o docs/coverage/coverage.html

# Apenas inventory:
go test -count=1 -covermode=atomic -coverprofile=docs/coverage/inventory.out ./internal/modules/inventory/...
go tool cover -func=docs/coverage/inventory.out | tail -1
```
