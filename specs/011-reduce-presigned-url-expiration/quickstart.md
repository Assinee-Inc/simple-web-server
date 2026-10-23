# Quickstart: Validação da Expiração de 60 Segundos das URLs Pré-assinadas

**Feature**: `011-reduce-presigned-url-expiration`  
**Date**: 2026-10-23  

---

## 1. Pré-requisitos e Ambiente

- Servidor backend rodando localmente (`make run` ou `go run cmd/web/main.go`) em `http://localhost:8080`.
- Frontend rodando localmente (`npm run dev --prefix frontend`) em `http://localhost:5173`.
- Bucket AWS S3 configurado com CORS ativo para `PUT` e `GET`.
- Token JWT de autenticação do criador.

---

## 2. Cenário 1: Upload Direto Dentro da Janela de 60 Segundos (Sucesso)

### Passo 1: Solicitar URL Pré-assinada de Upload
```bash
TOKEN="<seu_jwt_token>"

UPLOAD_RESP=$(curl -s -X POST http://localhost:8080/file/upload-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "teste-60s.pdf",
    "file_size": 1024,
    "content_type": "application/pdf"
  }')

echo "$UPLOAD_RESP" | jq .
```

**Resultado esperado:**
- `expires_in` é igual a `60`.
- `upload_url` contém o parâmetro `X-Amz-Expires=60`.

### Passo 2: Executar Upload Imediatamente (< 60s)
```bash
UPLOAD_URL=$(echo "$UPLOAD_RESP" | jq -r '.upload_url')
echo "%PDF-1.4 teste 60 segundos" > /tmp/teste-60s.pdf

curl -i -X PUT "$UPLOAD_URL" \
  -H "Content-Type: application/pdf" \
  --data-binary "@/tmp/teste-60s.pdf"
```

**Resultado esperado:**
- HTTP `200 OK` retornado pela AWS S3.

---

## 3. Cenário 2: Upload Direto Após Expiração de 60 Segundos (Rejeição)

### Passo 1: Solicitar URL e Aguardar 65 Segundos
```bash
EXPIRED_RESP=$(curl -s -X POST http://localhost:8080/file/upload-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "teste-expirado.pdf",
    "file_size": 1024,
    "content_type": "application/pdf"
  }')

EXPIRED_URL=$(echo "$EXPIRED_RESP" | jq -r '.upload_url')

echo "Aguardando 65 segundos para expiração..."
sleep 65
```

### Passo 2: Tentar Upload com a URL Expirada
```bash
curl -i -X PUT "$EXPIRED_URL" \
  -H "Content-Type: application/pdf" \
  --data-binary "@/tmp/teste-60s.pdf"
```

**Resultado esperado:**
- HTTP `403 Forbidden` retornado pela AWS S3 com XML contendo:
  ```xml
  <Error>
    <Code>AccessDenied</Code>
    <Message>Request has expired</Message>
    <Expires>...</Expires>
  </Error>
  ```

---

## 4. Cenário 3: Validação de Links de Download/Preview da Biblioteca

### Passo 1: Listar Arquivos da Biblioteca
```bash
FILES_RESP=$(curl -s -X GET "http://localhost:8080/files?page=1&limit=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$FILES_RESP" | jq '.files[0].s3_url'
```

**Resultado esperado:**
- O link `s3_url` contém o parâmetro `X-Amz-Expires=60`.
- Abrir a URL imediatamente (< 60s) faz o download/preview com HTTP `200 OK`.
- Abrir a mesma URL após 65 segundos retorna HTTP `403 Forbidden` (`Request has expired`).
