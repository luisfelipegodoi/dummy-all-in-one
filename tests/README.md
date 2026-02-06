# 🧪 System Tests – End-to-End (E2E)

Este diretório contém os **System Tests** do projeto, responsáveis por validar **cenários reais de execução**, incluindo:

- Criação de clusters Kind
- Instalação de infraestrutura. Ex: (Kind, Kaptain APIs, Crossplane, ArgoCD, NATS, LocalStack, DynamoDB, Registry, dentre outros)
- Deploy de aplicações
- Validações via HTTP, DynamoDB (AWS SDK), Kubernetes e Crossplane
- Atualização e remoção de recursos

---

## 🎯 Objetivo

Garantir que **fluxos completos de negócio** funcionem corretamente, validando:

- Infraestrutura provisionada corretamente
- Providers do Crossplane instalados e saudáveis
- Recursos criados/atualizados/removidos conforme esperado
- Aplicações acessíveis e funcionais
- Estados finais consistentes (ex: `Installed=true`, `Ready=true`, `pending=0`)

---

## 📁 Estrutura de Diretórios

```
tests/
└── system/
    ├── main_test.go          # Orquestrador principal (TestMain)
    ├── plan.go               # Mapeamento FLOW → clusters + infra
    ├── env.yaml              # Definição dos clusters e configurações globais
    ├── flows/                # Flows (cenários de negócio)
    │   ├── aws_only/
    │   │   ├── flow_test.go
    │   │   └── fixtures/
    │   ├── dynamodb_flow/
    │   │   ├── flow_test.go
    │   │   └── fixtures/
    │   └── scheduler_flow/
    │       ├── flow_test.go
    │       └── fixtures/
    ├── infra/
    │   ├── k8s/              # Manifests Kubernetes
    │   └── helm/             # Charts Helm usados nos testes
    ├── kubectl/              # Helpers kubectl + waits
    ├── helm/                 # Helpers Helm + waits
    └── utils/                # Helpers (DynamoDB, HTTP, retries, etc)
```

---

## 🔀 Conceito de **Flow**

Um **flow** representa um **cenário completo de system test**.

Exemplos:
- `aws_only`
- `dynamodb_flow`
- `scheduler_flow`

Cada flow define:
- Quais clusters serão criados
- Quais componentes serão instalados
- Quais testes serão executados

---

## 🌱 Selecionando um Flow

O flow é definido via variável de ambiente `FLOW`.

```bash
export FLOW=aws_only
```

O valor de `FLOW` é interpretado no arquivo `plan.go`, que retorna um **Plan** com os clusters e infra necessários.

---

## 🚀 Executando os Testes

```bash
export FLOW=aws_only
go test ./tests/system -count=1 -v
```

---

## 🧠 Como funciona o `TestMain`

O `TestMain` é o **cérebro do System Test**.

Fluxo de execução:

1. Resolver o **plan** baseado no `FLOW`
2. Criar apenas os clusters necessários
3. Para cada cluster:
   - Criar o cluster Kind
   - Instalar infraestrutura (Helm / kubectl)
   - Aguardar recursos ficarem prontos
4. Executar os testes (`m.Run()`)
5. (Opcional) Tear down dos clusters

---

## ⏳ Estratégia de Wait / Sincronização

Nenhum teste assume que algo está pronto imediatamente.

São utilizados waits explícitos para:

- Namespaces
- Deployments
- Jobs
- Providers do Crossplane
- CRDs
- Tabelas DynamoDB

---

## ➕ Criando um novo Flow

1. Criar diretório:
```
tests/system/flows/my_new_flow/
```

2. Atualizar `plan.go` e executar:
```bash
export FLOW=my_new_flow
go test ./tests/system -count=1 -v
```

---

## 🤝 Construído em Time

Made with ❤️ by Squad Armada
