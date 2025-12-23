#!/bin/bash
set -e  
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' 

# Pegar o diretório do script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
    echo -e "${GREEN}✅ Variáveis de ambiente carregadas do .env${NC}"
else
    echo -e "${RED}❌ Arquivo .env não encontrado em: ${ENV_FILE}${NC}"
    exit 1
fi

if [ -z "$POSTGRES_PORT_EXTERNAL" ]; then
    echo -e "${RED}❌ Variável POSTGRES_PORT_EXTERNAL não definida no .env${NC}"
    exit 1
fi


DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT_EXTERNAL}/${POSTGRES_DB}?sslmode=disable"
MIGRATIONS_PATH="./internal/database/migrations"

echo -e "${YELLOW}🔄 Iniciando migrations...${NC}"
echo -e "${YELLOW}📍 URL de conexão: postgres://${POSTGRES_USER}:***@${POSTGRES_HOST}:${POSTGRES_PORT_EXTERNAL}/${POSTGRES_DB}${NC}"

if command -v pg_isready &> /dev/null; then
    if ! PGPASSWORD=$POSTGRES_PASSWORD pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT_EXTERNAL" -U "$POSTGRES_USER" &> /dev/null; then
        echo -e "${RED}❌ Banco de dados não está acessível!${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  pg_isready não encontrado, verificando via Docker...${NC}"
    # Pegar o nome do container dinamicamente
    CONTAINER_NAME=$(docker ps --filter "ancestor=postgres:17" --format "{{.Names}}" | head -n 1)
    if [ -z "$CONTAINER_NAME" ]; then
        echo -e "${RED}❌ Container PostgreSQL não encontrado!${NC}"
        echo -e "${YELLOW}Execute: docker-compose up -d${NC}"
        exit 1
    fi
    if ! docker exec "$CONTAINER_NAME" pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" &> /dev/null; then
        echo -e "${RED}❌ Banco de dados não está acessível no container: ${CONTAINER_NAME}${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ Banco acessível!${NC}"

# Limpar variáveis de ambiente do PostgreSQL
unset PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

echo -e "${YELLOW}⬆️  Aplicando migrations...${NC}"
if migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" up; then
    echo -e "${GREEN}✅ Migrations aplicadas com sucesso!${NC}"
else
    echo -e "${RED}❌ Erro ao aplicar migrations${NC}"
    exit 1
fi

# Mostra versão atual
VERSION=$(migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" version 2>&1 | tail -n 1)
echo -e "${GREEN}📊 Versão atual: ${VERSION}${NC}"