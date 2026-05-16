# CI/CD — oficina-tech-ms-workshop

## Fluxo do Pipeline

```
push/PR → develop ou main
    │
    ▼
[ci.yml]  Lint + Vet + Tests + Coverage Gate (80%)
    │
    │  push para develop
    ▼
[release.yml]  Calcula versão → cria/atualiza PR release/* → main
    │
    │  PR release/* merged → main
    ▼
[deploy.yml]
    ├── Stage 1: Build & Test   (verifica compilação no main)
    ├── Stage 2: Docker         (build + push GHCR)
    ├── Stage 3: K8s            (valida secrets → infra-connect → k8s-provision → k8s-application → tag)
    ├── Stage 4: GitHub Release (publicado após tag confirmada em produção)
    └── Stage 5: SonarQube      (análise de qualidade)

[rollback.yml]  Redeployaponte versão anterior (workflow_dispatch manual)
```

## Workflows

| Arquivo | Trigger | Propósito |
|---------|---------|-----------|
| `ci.yml` | PR para `develop`/`main`, push para `develop` | Lint, vet, testes, coverage gate 80% |
| `release.yml` | Push para `develop`, `workflow_dispatch` | Cria/atualiza PR de release com versão calculada |
| `deploy.yml` | PR `release/*` → `main` merged, `workflow_dispatch` | Build, Docker, deploy K8s, tag, SonarQube |
| `rollback.yml` | `workflow_dispatch` manual | Redeployimagem de uma tag anterior sem criar nova tag |

## Versão automática (Conventional Commits)

| Tipo de commit | Bump |
|----------------|------|
| `feat:` | minor (0.X.0) |
| `feat!:` ou `BREAKING CHANGE` | major (X.0.0) |
| `fix:`, `chore:`, `docs:`, `refactor:` | patch (0.0.X) |

A tag é criada **somente após** o health check confirmar que o deploy está saudável.

## Variables (Settings → Secrets and Variables → Variables)

| Variável | Obrigatória | Descrição | Exemplo |
|----------|-------------|-----------|---------|
| `APP_NAME` | Sim | Nome do serviço no K8s | `ms-workshop` |
| `K8S_NAMESPACE` | Sim | Namespace Kubernetes | `app-oficina-tech` |
| `AWS_REGION` | Sim | Região AWS | `us-east-1` |
| `INFRA_TFSTATE_S3_PATH` | Sim | Path do TF state da infra | `s3://bucket/infra/terraform.tfstate` |
| `DB_TFSTATE_S3_PATH` | Sim | Path do TF state do banco | `s3://bucket/db/terraform.tfstate` |
| `SQS_INVENTORY_OP_REQUESTED_URL` | Sim | Fila SQS a consumir (← MS2) | `https://sqs.us-east-1.amazonaws.com/.../order-inventory-op-requested` |
| `SQS_INVENTORY_OP_SUCCEEDED_URL` | Sim | Fila SQS a publicar (→ MS2) | `https://sqs.us-east-1.amazonaws.com/.../order-inventory-op-succeeded` |
| `SQS_INVENTORY_OP_FAILED_URL` | Sim | Fila SQS a publicar (→ MS2) | `https://sqs.us-east-1.amazonaws.com/.../order-inventory-op-failed` |
| `SNS_INVENTORY_LOW_ALERT_ARN` | Sim | Tópico SNS para alertas de estoque baixo | `arn:aws:sns:us-east-1:...:inventory-low-alert` |
| `GO_VERSION` | Não | Versão do Go (default: `1.25`) | `1.25` |
| `GO_MAIN_PATH` | Não | Entrypoint (default: `cmd/api/main.go`) | `cmd/api/main.go` |
| `SERVER_PORT` | Não | Porta do serviço (default: `8083`) | `8083` |
| `HEALTH_ENDPOINT` | Não | Path do health check (default: `/health`) | `/health` |
| `SOURCE_BRANCH` | Não | Branch de origem do release (default: `develop`) | `develop` |
| `BASE_BRANCH` | Não | Branch de destino do release (default: `main`) | `main` |

## Secrets (Settings → Secrets and Variables → Secrets)

| Secret | Obrigatório | Descrição |
|--------|-------------|-----------|
| `AWS_ACCESS_KEY_ID` | Sim | Credencial AWS |
| `AWS_SECRET_ACCESS_KEY` | Sim | Credencial AWS |
| `AWS_SESSION_TOKEN` | Não | Credencial AWS (sessão temporária) |
| `DB_PASSWORD` | Sim | Senha do banco PostgreSQL (`db_ms3`) |
| `JWT_SECRET_KEY` | Sim | Secret compartilhado para assinatura JWT (mesmo valor em todos os MSs) |
| `GHCR_PAT` | Sim | Personal Access Token para o GitHub Container Registry |
| `DD_API_KEY` | Sim | Datadog API key |
| `DD_APP_KEY` | Sim | Datadog Application key |
| `DD_SITE` | Não | Datadog site (ex: `datadoghq.com`) |
| `SONAR_TOKEN` | Sim | Token do SonarQube |
| `SONAR_HOST_URL` | Sim | URL do servidor SonarQube |

## Como fazer rollback

```
Actions → Rollback Application → Run workflow
  version: v1.2.3  (tag existente com imagem no GHCR)
```

O rollback:
1. Valida formato da versão, existência da tag e da imagem no GHCR
2. Faz checkout **na tag** (manifests K8s corretos para aquela versão)
3. Lê infraestrutura **atual** do TF state no S3
4. Redeploya a imagem da tag para o cluster atual
5. **Não cria nova tag**

> **Atenção:** Operações de saga em andamento (`saga_operations` com status `PROCESSING`)
> no momento do rollback serão reprocessadas pelo consumer SQS ao subir — a tabela
> `saga_operations` garante idempotência via `saga_id + operation`.

## Composite Actions

```
.github/actions/
├── build/
│   ├── test/       Compila e roda go test
│   └── docker/     Build e push da imagem para GHCR
├── deploy/
│   ├── infra-connect/   Lê TF state S3, configura kubectl
│   ├── k8s-provision/   Aplica namespace, ConfigMap (SQS URLs, SNS ARNs), Secret (JWT, DB)
│   └── k8s-application/ Aplica Deployment, Service, aguarda health check
├── release/
│   ├── create-pr/       Calcula versão, cria branch release/*, abre PR
│   ├── update-pr/       Sincroniza develop na branch de release existente
│   ├── finalize-tag/    Cria a tag git após health check confirmado
│   └── publish/         Publica GitHub Release com changelog
└── quality/
    └── sonarqube/   Executa análise SonarQube
```
