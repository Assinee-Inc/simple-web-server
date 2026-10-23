# Feature Specification: Upload Direto para AWS S3 com URLs Pré-assinadas

**Feature Branch**: `direct_aws_file_upload` (Spec: `010-direct-aws-file-upload`)  
**Created**: 2026-10-23  
**Status**: Draft  
**Input**: User description: "O arquivo para upload não precisa passar pelo backend. O backend deve gerar uma url pre assinada para que o frontend consiga fazer o upload direto com a aws aliviando o servidor backend."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Obtenção de URL Pré-assinada para Upload (Priority: P1)

Um criador de conteúdo autenticado no Docffy deseja enviar um arquivo digital (ex.: PDF de e-book) para a biblioteca. Antes de realizar o upload do arquivo binário, o cliente (frontend) solicita ao backend uma URL pré-assinada de upload informando os metadados do arquivo (nome original, tamanho em bytes e tipo de conteúdo). O backend valida os metadados, assegura o isolamento de namespace por criador (`files/{creator_id}/...`) e gera uma URL pré-assinada do AWS S3 (`PUT`) com tempo de expiração determinado, retornando a URL e a chave S3 segura para o frontend.

**Why this priority**: É a fundação do desacoplamento de tráfego binário, permitindo que o frontend envie bytes diretamente para o S3 sem sobrecarregar a memória, CPU e conexões de rede do backend.

**Independent Test**: Pode ser testado autenticando como criador e enviando uma requisição `POST /file/upload-url` com dados válidos de um arquivo. O backend deve responder com status `200 OK`, uma URL pré-assinada contendo assinatura SigV4 da AWS e a chave (`key`) isolada no diretório do criador.

**Acceptance Scenarios**:

1. **Given** um criador autenticado e com cadastro ativo, **When** ele solicita uma URL pré-assinada informando um nome de arquivo válido (`manual.pdf`), tamanho permitido (<= 100MB) e content type (`application/pdf`), **Then** o backend responde `200 OK` com a URL pré-assinada (`upload_url`), método HTTP `PUT`, tempo de expiração e a chave única S3 (`key`) estruturada como `files/{creator_id}/{unique_prefix}-{safe_filename}`.
2. **Given** uma requisição de geração de URL com extensão ou tipo não suportado, **When** o criador submete a requisição, **Then** o backend rejeita com erro `400 Bad Request` indicando que o tipo de arquivo não é suportado.
3. **Given** uma requisição cujo tamanho declarado ultrapasse o limite de 100MB, **When** o criador submete a requisição, **Then** o backend rejeita com `400 Bad Request` informando o excesso de tamanho.
4. **Given** um usuário não autenticado ou sem onboarding de criador concluído, **When** tenta solicitar uma URL pré-assinada, **Then** o backend recusa com `401 Unauthorized` ou `403 Forbidden`.

---

### User Story 2 - Upload Direto do Navegador para a AWS S3 (Priority: P1)

Após obter a URL pré-assinada, o frontend realiza o upload do binário (arquivo selecionado pelo usuário) diretamente para o bucket da AWS via requisição HTTP `PUT`, sem trafegar pelo servidor Go. O navegador se comunica com a AWS S3 através de CORS devidamente configurado, transmitindo o payload binário de forma otimizada com feedback visual de progresso para o usuário.

**Why this priority**: Elimina 100% da carga de tráfego de arquivos pesados no backend, evitando timeouts de gateway/proxy (ex.: timeout de 30s do Heroku) e consumo desnecessário de largura de banda do servidor.

**Independent Test**: Pode ser testado executando um upload via `curl -X PUT --upload-file <path> <upload_url>` ou pelo frontend e verificando resposta `200 OK` da AWS S3 com o arquivo persistido no bucket.

**Acceptance Scenarios**:

1. **Given** uma URL pré-assinada válida emitida pelo backend, **When** o frontend envia o binário via HTTP `PUT` diretamente para a AWS S3 com o `Content-Type` acordado, **Then** a AWS S3 aceita o upload e retorna código de status `200 OK`.
2. **Given** uma URL expirada ou com cabeçalhos divergentes dos assinados, **When** o upload é enviado diretamente para a AWS, **Then** a AWS S3 recusa a requisição com `403 Forbidden`.

---

### User Story 3 - Confirmação e Registro de Metadados do Arquivo no Backend (Priority: P1)

Uma vez concluído o upload diretamente na AWS S3 com sucesso (`200 OK`), o frontend notifica o backend enviando os metadados finais do arquivo (chave S3 recebida, nome de exibição, descrição opcional, tamanho e nome original). O backend valida que a chave informada pertence ao criador autenticado, valida a existência do objeto no S3, persiste o registro na tabela de arquivos da biblioteca e retorna o registro criado em JSON.

**Why this priority**: Garante que o banco de dados permaneça consistente apenas com arquivos cujo upload tenha sido finalizado com sucesso no S3, prevenindo registros órfãos e tentativas de apropriação indevida de arquivos entre criadores.

**Independent Test**: Pode ser testado após o upload no S3, enviando `POST /file/complete-upload` com a chave gerada e metadados. O backend deve responder com `201 Created` e o arquivo deve constar na listagem `GET /file`.

**Acceptance Scenarios**:

1. **Given** que o arquivo foi enviado com sucesso para o S3, **When** o frontend chama `POST /file/complete-upload` com a `key` válida pertencente ao criador, **Then** o backend registra o arquivo no banco de dados e retorna `201 Created` com os detalhes do arquivo (`public_id`, `name`, `original_name`, `size`, `s3_key`, etc.).
2. **Given** uma tentativa de confirmação onde a `key` não pertence ao prefixo do criador autenticado (`files/{outro_creator_id}/...`), **When** a requisição é processada, **Then** o backend recusa com `403 Forbidden` e impede a persistência no banco.
3. **Given** uma requisição de confirmação com dados incompletos ou inválidos, **When** submetida, **Then** o backend rejeita com `400 Bad Request`.

---

### User Story 4 - Atualização da Interface do Usuário (Frontend SPA) (Priority: P2)

Na tela de gerenciamento de arquivos (`FileLibraryPage.tsx`), o formulário de envio de arquivo passa a orquestrar o fluxo em duas etapas (obtenção de URL pré-assinada -> upload direto para o S3 com PUT -> confirmação no backend), mantendo uma experiência limpa com estados de carregamento ("Preparando upload...", "Enviando para a nuvem...", "Finalizando...") e notificações de sucesso ou erro amigáveis via toast.

**Why this priority**: Torna o novo fluxo transparente e intuitivo para o criador de conteúdo, garantindo alta confiabilidade e feedback visual em tempo real.

**Independent Test**: Testar a submissão de arquivo pela interface SPA e validar no console/network do navegador que a requisição de upload pesado é disparada para `*.s3.amazonaws.com` e não para o host do backend.

**Acceptance Scenarios**:

1. **Given** o criador na página da biblioteca de arquivos, **When** ele seleciona um arquivo PDF e clica em enviar, **Then** a interface executa a requisição da URL pré-assinada, transfere o arquivo diretamente para o S3 e confirma o cadastro no backend, atualizando a listagem de arquivos sem recarregar a página.

---

### Edge Cases

- **Upload abandonado ou falha de rede para o S3**: Se o criador cancelar o upload ou perder conexão antes do fim do envio para o S3, o backend não recebe a chamada de confirmação e nenhum registro corrompido ou incompleto é inserido no banco de dados. Políticas de ciclo de vida do bucket S3 podem ser configuradas para expirar chaves não confirmadas se necessário.
- **Divergência de Content-Type**: O `Content-Type` enviado pelo frontend no `PUT` para a AWS deve ser idêntico ao informado na solicitação da URL pré-assinada; divergências devem ser evitadas padronizando a detecção do tipo MIME no cliente.
- **Manipulação de Key por atacante**: Um usuário malicioso que tente enviar um `key` de outro criador no endpoint de confirmação deve ser imediatamente barrado pela validação server-side do prefixo `files/{creator_id}/`.
- **Compatibilidade Retroativa**: O endpoint legado `POST /file/upload` via `multipart/form-data` pode ser mantido funcional ou deprecado de forma segura, garantindo que integrações existentes não quebrem abruptamente.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O backend DEVE expor um endpoint autenticado `POST /file/upload-url` (ou `/file/presign-upload`) que receba metadados do arquivo (`file_name`, `file_size`, `content_type`) e retorne uma URL pré-assinada HTTP `PUT` do AWS S3, a chave `key` de destino e os cabeçalhos obrigatórios.
- **FR-002**: A chave S3 (`key`) gerada pelo backend DEVE conter obrigatoriamente o identificador do criador autenticado (`files/{creator_id}/{unique_id}-{safe_name}`) para garantir isolamento multi-tenant estrito.
- **FR-003**: O backend DEVE validar o tamanho máximo permitido (100MB) e os tipos de arquivo permitidos (PDF e tipos suportados pela plataforma) antes de gerar a URL pré-assinada.
- **FR-004**: O backend DEVE expor o método `GeneratePresignedPutURL(key, contentType string, expirationSeconds int) (string, error)` na interface `S3Storage` (`pkg/storage`), utilizando o cliente de pré-assinatura do AWS SDK v2 (`PresignPutObject`).
- **FR-005**: O backend DEVE expor um endpoint autenticado `POST /file/complete-upload` que receba os metadados do arquivo enviado (`key`, `name`, `description`, `original_name`, `size`, `content_type`), valide o pertencimento da chave ao criador autenticado e salve o registro `File` no banco de dados.
- **FR-006**: O frontend React (`FileLibraryPage.tsx`) DEVE deixar de enviar o binário via `multipart/form-data` para o endpoint `/file/upload` do backend e passar a executar o fluxo:
  1. Solicitação da URL pré-assinada (`POST /file/upload-url`).
  2. Upload do binário direto para a AWS S3 via `fetch(upload_url, { method: 'PUT', body: file, headers: { 'Content-Type': contentType } })`.
  3. Confirmação e registro via `POST /file/complete-upload`.
- **FR-007**: As respostas do backend DEVEM seguir o padrão REST JSON já estabelecido no projeto (`{"success": true, ...}` ou `{"success": false, "code": "...", "error": "..."}`).

### Key Entities

- **File**: Registro da biblioteca de arquivos pertencente a um `Creator` (`creator_id`). Contém `public_id`, `name`, `original_name`, `description`, `file_type`, `s3_key`, `s3_url`, `size` e status ativo/inativo.
- **UploadRequestDTO**: Objeto de requisição para obter a URL pré-assinada (`file_name`, `file_size`, `content_type`).
- **UploadResponseDTO**: Objeto retornado pelo backend contendo `upload_url`, `key`, `method`, `expires_in` e headers necessários.
- **CompleteUploadDTO**: Objeto de confirmação enviado pelo frontend após o sucesso do upload no S3 (`key`, `name`, `description`, `original_name`, `size`, `content_type`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 0 bytes do binário do arquivo trafegam pelo servidor backend Go durante o upload realizado pelo frontend.
- **SC-002**: Redução de tempo de resposta da API de upload de arquivos no backend para < 200ms (tempo apenas para emitir a URL pré-assinada e confirmar metadados).
- **SC-003**: 100% dos uploads concluídos pelo frontend geram registros íntegros no banco de dados associados ao criador correto.
- **SC-004**: Tentativas de confirmação de arquivos com chave S3 fora do namespace do criador são rejeitadas com 100% de precisão (403 Forbidden).
