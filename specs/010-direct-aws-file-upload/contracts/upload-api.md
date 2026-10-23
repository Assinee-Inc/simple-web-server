# Contracts: Upload Direto para AWS S3 com URLs Pré-assinadas

**Feature**: `010-direct-aws-file-upload`  
**Base Path**: `/file`  
**Authentication**: `Authorization: Bearer <JWT>` + Perfil de Criador Ativo

---

## 1. Solicitar URL Pré-assinada de Upload

Solicita uma URL pré-assinada do AWS S3 para permitir o upload direto do navegador.

- **Method**: `POST`
- **Path**: `/file/upload-url`
- **Headers**:
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`

### Request Body
```json
{
  "file_name": "guia-definitivo.pdf",
  "file_size": 2048576,
  "content_type": "application/pdf"
}
```

### Responses

#### `200 OK` (URL Gerada com Sucesso)
```json
{
  "success": true,
  "upload_url": "https://docffy-bucket.s3.us-east-1.amazonaws.com/files/42/guia-definitivo-a1b2c3d4.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&...",
  "key": "files/42/guia-definitivo-a1b2c3d4.pdf",
  "method": "PUT",
  "headers": {
    "Content-Type": "application/pdf"
  },
  "expires_in": 900
}
```

#### `400 Bad Request` (Validação de Metadados)
```json
{
  "success": false,
  "code": "INVALID_FILE",
  "error": "Apenas arquivos PDF são permitidos ou tamanho excede 100MB"
}
```

#### `401 Unauthorized`
```json
{
  "success": false,
  "code": "UNAUTHORIZED",
  "error": "Usuário não autenticado"
}
```

#### `403 Forbidden`
```json
{
  "success": false,
  "code": "CREATOR_ONBOARDING_REQUIRED",
  "error": "Conclua o cadastro de criador para acessar este recurso"
}
```

---

## 2. Upload Direto para a AWS S3 (Cliente -> AWS)

Executa o envio do binário do arquivo diretamente para o bucket da AWS S3.

- **Method**: `PUT`
- **URL**: `<upload_url>` (retornada no endpoint anterior)
- **Headers**:
  - `Content-Type: application/pdf` (deve corresponder ao solicitado)
- **Body**: Binário do arquivo (Blob / File / Buffer)

### Responses (AWS S3)

#### `200 OK`
Upload concluído com sucesso no S3. O corpo da resposta é vazio e inclui o cabeçalho `ETag`.

#### `403 Forbidden`
URL expirada, assinatura inválida ou cabeçalhos divergentes dos assinados.

---

## 3. Confirmar e Registrar Upload Concluído

Registra o arquivo na biblioteca do criador após confirmação de upload no S3.

- **Method**: `POST`
- **Path**: `/file/complete-upload`
- **Headers**:
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`

### Request Body
```json
{
  "key": "files/42/guia-definitivo-a1b2c3d4.pdf",
  "name": "Guia Definitivo do Docffy",
  "description": "Material complementar do e-book principal",
  "original_name": "guia-definitivo.pdf",
  "size": 2048576,
  "content_type": "application/pdf"
}
```

### Responses

#### `201 Created` (Arquivo Registrado na Biblioteca)
```json
{
  "success": true,
  "message": "Arquivo registrado com sucesso!",
  "file": {
    "id": 105,
    "public_id": "fil_987654321",
    "name": "Guia Definitivo do Docffy",
    "original_name": "guia-definitivo.pdf",
    "description": "Material complementar do e-book principal",
    "file_type": "pdf",
    "s3_key": "files/42/guia-definitivo-a1b2c3d4.pdf",
    "s3_url": "https://docffy-bucket.s3.us-east-1.amazonaws.com/files/42/guia-definitivo-a1b2c3d4.pdf",
    "size": 2048576,
    "creator_id": 42,
    "is_active": true,
    "created_at": "2026-10-23T10:30:00Z",
    "updated_at": "2026-10-23T10:30:00Z"
  }
}
```

#### `403 Forbidden` (Violação de Namespace do Criador)
```json
{
  "success": false,
  "code": "FORBIDDEN",
  "error": "A chave do arquivo não pertence ao criador autenticado"
}
```

#### `400 Bad Request` (Dados Inválidos)
```json
{
  "success": false,
  "code": "INVALID_PAYLOAD",
  "error": "Nome do arquivo e chave S3 são obrigatórios"
}
```
