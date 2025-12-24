# 💳 Payments - Sistema de Gerenciamento de Transações

Sistema para processamento de transações financeiras com garantia de idempotência usando Redis e PostgreSQL.

## 🚀 Tecnologias

- **Go 1.21+** - Linguagem principal
- **PostgreSQL 17** - Banco de dados relacional
- **Redis Stack** - Cache e controle de idempotência
- **Docker & Docker Compose** - Containerização
- **Gin** - Framework HTTP
- **golang-migrate** - Gerenciamento de migrations

## 📋 Pré-requisitos

- Docker & Docker Compose
- Go 1.21+
- golang-migrate CLI

```bash
# Instalar golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## ⚙️ Configuração

### 1. Clone e configure

```bash
git clone https://github.com/seu-usuario/payments.git
cd payments
```

### 2. Crie o arquivo `.env`

```env
POSTGRES_HOST=localhost
POSTGRES_USER=admin12313
POSTGRES_PASSWORD=20242024
POSTGRES_DB=payments
POSTGRES_PORT_EXTERNAL=5440

API_PORT=8080

REDIS_PASSWORD=senhaforte124
REDIS_PORT=6388
```

### 3. Inicie os serviços

```bash
docker-compose up -d
```

### 4. Execute as migrations

```bash
chmod +x run-migrations.sh
./run-migrations.sh
```

## 🏗️ Arquitetura

O projeto segue **Clean Architecture** com separação clara de responsabilidades:

```
cmd/
└── api/main.go              # Entry point da aplicação

internal/
├── cache/                   # Camada de cache (Redis)
│   └── idempotency.go       # Controle de duplicação
│
├── controllers/             # Camada HTTP (handlers)
│   └── transactions.go      # Endpoints REST
│
├── usecases/                # Camada de Lógica de Negócio
│   └── transactions.go      # Regras de transferência
│
├── repository/              # Camada de Acesso a Dados
│   └── transaction_repository.go
│
├── models/                  # Entidades de Domínio
│   ├── account.go
│   └── transaction.go
│
└── database/
    ├── postgres/            # Conexão PostgreSQL
    ├── redis/               # Conexão Redis
    └── migrations/          # SQL migrations
```

### Fluxo de Requisição

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP Request
       ▼
┌─────────────────┐
│   Controller    │  ← Recebe requisição, valida JSON
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    UseCase      │  ← Lógica de negócio + Redis (idempotência)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Repository    │  ← Acessa PostgreSQL
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PostgreSQL    │
└─────────────────┘
```

### Componentes

- **Controllers**: Recebem requisições HTTP e retornam respostas
- **UseCases**: Contém a lógica de negócio (validações, regras)
- **Repository**: Interface com o banco de dados
- **Cache**: Controle de idempotência via Redis
- **Models**: Estruturas de dados (entidades)

## 🔌 API

### POST /api/v1/transactions

Cria uma nova transação.

**Request:**
```json
{
  "external_id": "TXN-2025-001",
  "from_account_id": "uuid-origem",
  "to_account_id": "uuid-destino",
  "type": "TRANSFER",
  "amount": 15075,
  "currency": "BRL",
  "status": "PENDING"
}
```

**Responses:**
- `201` - Transação criada
- `409` - Transação duplicada
- `400` - Dados inválidos
- `500` - Erro interno

## 🔐 Idempotência

O sistema previne transações duplicadas usando Redis:

1. Cliente envia `external_id` único
2. Redis verifica se já existe
3. Se existe: retorna `409 Conflict`
4. Se não existe: processa e salva no PostgreSQL

**TTL do Cache:**
- `PROCESSING`: 30 segundos
- `COMPLETED`: 10 segundos

## 📊 Banco de Dados

### Tabela: account
```sql
id, user_name, user_cpf_cnpj, blocked, user_email, created_at
```

### Tabela: transactions
```sql
id, external_id, from_account_id, to_account_id, 
type, amount, currency, status, created_at
```

**Índices criados:**
- `from_account_id` (origem)
- `to_account_id` (destino)
- `status` (filtros)
- `created_at` (ordenação)
- Composto: `(from_account_id, status)`

## 🛠️ Comandos Úteis

```bash
# Iniciar
docker-compose up -d

# Parar
docker-compose down

# Logs
docker-compose logs -f api

# Migrations
./run-migrations.sh

# Reverter última migration
migrate -path ./internal/database/migrations -database "$DB_URL" down 1
```

## 🐛 Troubleshooting

**Erro: Port already in use**
```bash
docker-compose down
sudo lsof -ti:8080 | xargs kill -9
```

**Erro: Dirty database**
```bash
migrate -path ./internal/database/migrations -database "$DB_URL" force 1
```

## 📝 Variáveis de Ambiente

| Variável | Descrição |
|----------|-----------|
| `POSTGRES_HOST` | Host do PostgreSQL |
| `POSTGRES_USER` | Usuário do banco |
| `POSTGRES_PASSWORD` | Senha do banco |
| `POSTGRES_DB` | Nome do banco |
| `POSTGRES_PORT_EXTERNAL` | Porta externa |
| `API_PORT` | Porta da API |
| `REDIS_PASSWORD` | Senha do Redis |
| `REDIS_PORT` | Porta do Redis |

## 📄 Licença

MIT License

---

**Documentação da API (Swagger)**: Em breve
