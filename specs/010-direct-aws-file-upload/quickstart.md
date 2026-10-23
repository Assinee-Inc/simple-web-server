# Quickstart: Testando Upload Direto para AWS S3

**Feature**: `010-direct-aws-file-upload`  
**Date**: 2026-10-23

---

## 1. Pré-requisitos e Ambiente

- Servidor backend rodando localmente (`make run` ou `go run cmd/web/main.go`) em `http://localhost:8080`.
- Frontend rodando localmente (`npm run dev --prefix frontend`) em `http://localhost:5173`.
- Variáveis de ambiente configuradas no `.env`:
  ```bash
  S3_BUCKET_NAME=seu-bucket-s3
  S3_REGION=us-east-1
  S3_ACCESS_KEY=...
  S3_SECRET_KEY=...
  ```
- Configuração de CORS no Bucket S3 (AWS Console -> S3 -> Bucket -> Permissions -> Cross-origin resource sharing):
  ```json
  [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "PUT", "HEAD"],
      "AllowedOrigins": ["http://localhost:5173", "http://localhost:8080", "*"],
      "ExposeHeaders": ["ETag"]
    }
  ]
  ```

---

## 2. Testando o Fluxo via cURL

### Passo 1: Solicitar URL Pré-assinada
```bash
TOKEN="<seu_jwt_token>"

UPLOAD_RESP=$(curl -s -X POST http://localhost:8080/file/upload-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "teste.pdf",
    "file_size": 1024,
    "content_type": "application/pdf"
  }')

echo "$UPLOAD_RESP"
UPLOAD_URL=$(echo "$UPLOAD_RESP" | jq -r '.upload_url')
KEY=$(echo "$UPLOAD_RESP" | jq -r '.key')
```

### Passo 2: Upload Direto para o AWS S3 (sem passar pelo backend)
```bash
# Cria um arquivo temporário de teste
echo "%PDF-1.4 teste de upload direto" > /tmp/teste.pdf

# Envia via PUT direto para a AWS
curl -v -X PUT "$UPLOAD_URL" \
  -H "Content-Type: application/pdf" \
  --data-binary "@/tmp/teste.pdf"
```

A AWS S3 deve retornar `HTTP/1.1 200 OK`.

### Passo 3: Confirmar e Registrar Arquivo no Backend
```bash
curl -s -X POST http://localhost:8080/file/complete-upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"key\": \"$KEY\",
    \"name\": \"Arquivo de Teste Direto\",
    \"description\": \"Enviado diretamente para o S3\",
    \"original_name\": \"teste.pdf\",
    \"size\": 1024,
    \"content_type\": \"application/pdf\"
  }" | jq .
```

O arquivo aparecerá registrado na biblioteca do criador!

---

## 3. Testando via Interface do Usuário (Frontend)

1. Faça login na aplicação frontend em `http://localhost:5173`.
2. Acesse a página **Biblioteca de Arquivos** (`/files`).
3. Abra as Ferramentas do Desenvolvedor do navegador (F12) na aba **Rede (Network)**.
4. Clique em **Enviar Arquivo**, preencha o nome, selecione um PDF e envie.
5. Inspecione as requisições de rede:
   - Requisição 1: `POST /file/upload-url` para o backend Go (payload leve em JSON, ~200ms).
   - Requisição 2: `PUT https://...s3.amazonaws.com/...` diretamente para o bucket da AWS S3 com o payload binário do arquivo.
   - Requisição 3: `POST /file/complete-upload` para o backend Go (confirmação em JSON, ~150ms).
6. Confirme que o arquivo é listado na tabela da biblioteca com sucesso.
