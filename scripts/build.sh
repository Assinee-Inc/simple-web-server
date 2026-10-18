#!/bin/bash

# Script de build para o projeto Docffy
set -e

echo "🚀 Iniciando build do Docffy..."

# Criar diretório bin se não existir
mkdir -p bin

# Compilar o projeto
echo "📦 Compilando o projeto..."
GOOS=linux GOARCH=amd64 go build -o bin/docffy cmd/web/main.go

# Tornar o binário executável
chmod +x bin/docffy

echo "✅ Build concluído! Binário criado em: bin/docffy"
echo "📊 Tamanho do binário: $(du -h bin/docffy | cut -f1)"