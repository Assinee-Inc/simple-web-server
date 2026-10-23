# Feature Specification: Autenticação via Google e Migração Completa para APIs (Sem Senhas)

**Feature Branch**: `009-google-auth-api-migration`  
**Created**: 2026-10-19  
**Status**: Draft  
**Input**: User description: "O Docffy é um sistema de que nao precisa se preocupar com as senhas dos usuários. A ideia é implementar a autenticacao via google e deixar de armazenar senhas no sistema. O sistema deverá deixar de ser SSR e passará a responder via APIs."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Autenticação via Conta Google Sem Senhas (Priority: P1)

Um usuário (novo ou já existente na plataforma) acessa o Docffy utilizando exclusivamente sua conta Google, sem precisar criar, digitar, armazenar ou recuperar senhas. Ao se autenticar, o usuário recebe uma credencial de acesso baseada em token para consumir os recursos protegidos da API, e contas já existentes com o mesmo endereço de e-mail são vinculadas automaticamente à identidade Google.

**Why this priority**: Substitui o mecanismo central de autenticação do sistema, elimina o armazenamento de senhas e habilita a comunicação stateless via API para todos os módulos autenticados.

**Independent Test**: Pode ser testado autenticando uma conta Google (tanto de um e-mail já existente na base quanto de um novo e-mail) através da API, verificando a emissão do token de acesso, o vínculo automático da conta existente sem uso de senha e a consulta aos dados do usuário autenticado.

**Acceptance Scenarios**:

1. **Given** um usuário já cadastrado anteriormente no Docffy com o mesmo e-mail de sua conta Google, **When** ele se autentica via Google na API pela primeira vez, **Then** o sistema vincula automaticamente sua conta ao identificador único do Google, marca seu e-mail como verificado e retorna um token de acesso válido junto com o estado do seu perfil.
2. **Given** um usuário que nunca utilizou o Docffy, **When** ele se autentica com uma conta Google válida e com e-mail verificado, **Then** o sistema cria imediatamente seu registro de usuário (sem senha e com e-mail já verificado) e retorna um token de acesso indicando que o onboarding de criador está pendente.
3. **Given** uma tentativa de autenticação com credencial Google inválida, expirada ou cujo e-mail não esteja verificado pelo Google, **When** a requisição é enviada à API, **Then** o sistema recusa a autenticação com código de status de não autorizado e mensagem de erro estruturada.
4. **Given** um usuário autenticado, **When** ele acessa o endpoint de perfil/sessão atual ou solicita encerramento da sessão, **Then** o sistema retorna os dados atualizados do usuário/criador ou encerra o acesso de acordo com a operação solicitada.

---

### User Story 2 - Onboarding de Criador em Duas Etapas via API (Priority: P2)

Após entrar pela primeira vez com sua conta Google (Etapa 1), o novo usuário completa seu cadastro comercial como criador (Etapa 2) enviando via API seus dados obrigatórios (CPF, data de nascimento, telefone, nome social opcional e aceite dos termos de uso). Enquanto essa segunda etapa não for concluída, o usuário permanece autenticado, mas o acesso às funcionalidades exclusivas de criador (gestão de ebooks, arquivos, vendas e clientes) fica condicionado à conclusão do perfil de criador.

**Why this priority**: Preserva a experiência fluida de entrada via Google ("Continue com Google") sem abrir mão das validações fiscais e legais obrigatórias para criadores no Docffy.

**Independent Test**: Pode ser testado autenticando um novo usuário via Google, verificando que chamadas às rotas de negócio de criador informam pendência de onboarding, submetendo os dados de criador no endpoint de onboarding e confirmando a liberação do acesso às funcionalidades de criador.

**Acceptance Scenarios**:

1. **Given** um usuário autenticado via Google que ainda não possui perfil de criador vinculado, **When** ele envia CPF válido, data de nascimento comprovando maioridade (18+ anos), telefone e aceite dos termos via API, **Then** o sistema cria o perfil de criador vinculado ao usuário, inicia o provisionamento da conta de recebimento e retorna os dados do criador criado.
2. **Given** um usuário autenticado que ainda não concluiu a Etapa 2 (perfil de criador), **When** ele tenta acessar endpoints exclusivos de operação de criador (como criação de ebooks, arquivos ou consulta de vendas), **Then** a API bloqueia a operação com resposta estruturada informando que o cadastro de criador precisa ser concluído primeiro.
3. **Given** dados inválidos no onboarding (CPF inválido ou já em uso, menor de 18 anos, ou termos não aceitos), **When** o usuário submete o cadastro de criador via API, **Then** a API rejeita a requisição com erros de validação detalhados por campo.

---

### User Story 3 - Migração Completa de Rotas SSR do Criador e Fluxos Públicos para APIs (Priority: P3)

Todas as operações do sistema — tanto a área autenticada do criador (dashboard, configurações, ebooks, arquivos, clientes, vendas, transações e onboarding de recebimentos) quanto os fluxos públicos de compradores (consulta de página de vendas, checkout e download de produtos adquiridos) — passam a responder exclusivamente via contratos de API estruturados, descontinuando a renderização de páginas HTML no servidor (SSR).

**Why this priority**: Conclui a transição arquitetural do Docffy de uma aplicação monolítica SSR para um provedor de APIs desacoplado da camada de apresentação visual.

**Independent Test**: Pode ser testado executando chamadas para os endpoints de cada domínio (conta, biblioteca, vendas, entrega e assinatura) e verificando que 100% das respostas de sucesso, validação e erro são retornadas em formato estruturado de API (JSON), sem renderização de templates HTML ou redirecionamentos de sessão baseados em formulários web.

**Acceptance Scenarios**:

1. **Given** um criador autenticado e com onboarding concluído, **When** ele realiza operações de listagem, criação, visualização, atualização ou exclusão nos módulos de conta/dashboard, biblioteca (ebooks e arquivos), vendas (clientes, compras e transações) e assinatura, **Then** todas as operações recebem e retornam dados estruturados via API com códigos de status apropriados.
2. **Given** um comprador acessando uma oferta pública de ebook, realizando checkout ou baixando seu arquivo adquirido, **When** o cliente consumidor consulta os endpoints públicos de página de vendas, checkout e entrega, **Then** o sistema retorna os dados da oferta, processa o checkout e entrega o arquivo protegido com marca d'água via contratos de API.
3. **Given** qualquer requisição inválida, não autenticada ou não autorizada em qualquer módulo do sistema, **When** o servidor responde ao cliente, **Then** a resposta contém código de status semântico e payload estruturado de erro, sem redirecionar para páginas HTML de login ou erro.

---

### Edge Cases

- **Conta Google sem e-mail verificado**: Se a conta Google retornar `email_verified = false`, a autenticação deve ser recusada imediatamente com mensagem clara.
- **Rotas legadas de senha e confirmação manual de e-mail**: Requisições para antigos fluxos de login por senha, esqueci minha senha, redefinição de senha e confirmação manual de e-mail por token devem ser removidas (retornando recurso não encontrado).
- **Conflito de CPF no onboarding em 2 etapas**: Se um novo usuário autenticado via Google tentar concluir o onboarding informando um CPF já vinculado a outro criador, a API deve rejeitar a solicitação com erro de conflito sem corromper o registro do usuário autenticado.
- **Indisponibilidade temporária do provedor Google ou serviços externos (Receita Federal / Gateways de Pagamento)**: A API deve responder com erro estruturado adequado sem expor detalhes internos da infraestrutura.
- **Usuário existente com conta Google já vinculada a outro ID**: O sistema deve garantir unicidade tanto do identificador Google quanto do e-mail, impedindo que duas contas distintas compartilhem a mesma identidade Google.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema DEVE autenticar usuários exclusivamente por meio de identidade Google verificada, emitindo um token de acesso stateless para consumo das rotas protegidas da API.
- **FR-002**: O sistema DEVE remover completamente o armazenamento e o processamento de senhas, hashes de senha, tokens de redefinição de senha, timestamps de redefinição de senha e tokens de verificação manual de e-mail do modelo de usuário e do banco de dados.
- **FR-003**: O sistema DEVE remover todas as rotas, fluxos e envios de e-mail relacionados a login por senha, recuperação de senha ("esqueci minha senha"), redefinição de senha e confirmação manual de conta por e-mail.
- **FR-004**: O sistema DEVE vincular automaticamente usuários já existentes na base de dados ao seu identificador único do Google quando o endereço de e-mail autenticado pelo Google coincidir com o e-mail já cadastrado no sistema.
- **FR-005**: Para novos e-mails autenticados via Google, o sistema DEVE realizar o cadastro em duas etapas:
  1. **Etapa 1 (Autenticação Google)**: Criação automática do registro `User` com nome, e-mail, identificador Google e e-mail marcado como verificado, retornando o token de acesso e a indicação de que o perfil `Creator` está pendente.
  2. **Etapa 2 (Onboarding de Criador)**: Endpoint autenticado que recebe CPF, data de nascimento, telefone, nome social (opcional) e aceite dos termos de uso para validar e criar o perfil `Creator` vinculado ao `User`.
- **FR-006**: O sistema DEVE impedir que usuários autenticados que ainda não concluíram a Etapa 2 (criação do perfil `Creator`) executem operações exclusivas de criador, retornando uma resposta estruturada indicando pendência de conclusão do perfil.
- **FR-007**: O sistema DEVE substituir 100% das respostas baseadas em renderização de templates HTML no servidor (SSR, flash messages em sessão e redirecionamentos de formulários web) por endpoints e respostas estruturadas de API (JSON) em todos os módulos do sistema (`auth`, `account`, `library`, `sales`, `delivery`, `subscription` e `shared`).
- **FR-008**: Os módulos da área autenticada do criador DEVEM expor via API todas as capacidades atualmente existentes em SSR:
  - **Conta e Dashboard**: resumo do dashboard, consulta/atualização de configurações (nome social, identificador de pixel de rastreamento) e gerenciamento de status/onboarding da conta de recebimento.
  - **Biblioteca**: listagem, criação, detalhamento, edição e remoção de ebooks e arquivos, associação/desassociação de arquivos a ebooks e obtenção de dados de pré-visualização da página de vendas.
  - **Vendas e Clientes**: listagem e exportação CSV de clientes, listagem de vendas realizadas, bloqueio/desbloqueio de download, reenvio de link de download e detalhamento de transações.
- **FR-009**: Os fluxos públicos voltados ao comprador DEVEM operar via API:
  - Consulta pública dos dados da página de vendas e de checkout de um ebook.
  - Validação de comprador, criação de sessão/cobrança de pagamento e consulta de confirmação da compra.
  - Download do arquivo digital protegido com marca d'água dinâmica mediante hash válido de compra.
- **FR-010**: O sistema DEVE preservar integralmente as regras de negócio e princípios constitucionais existentes em todas as operações de API: proteção de conteúdo com marca d'água dinâmica e limites de download, validação de webhooks e idempotência de pagamentos, validação de CPF, verificação de maioridade (18+ anos) e sanitização de entradas.

### Key Entities *(include if feature involves data)*

- **User (Usuário)**: Representa a identidade autenticada na plataforma. Contém identificador público (`public_id`), nome de exibição (`username`), e-mail único (`email`), identificador único da conta Google (`google_id`), data de verificação do e-mail (`email_verified_at`) e data de aceite dos termos (`terms_accepted_at`). Não possui senha nem tokens de recuperação/verificação manual.
- **Creator (Criador)**: Representa o perfil comercial e fiscal vinculado a um `User` (`user_id`), criado na segunda etapa do onboarding com nome completo, nome social, CPF validado, data de nascimento (18+ anos), telefone e dados de integração de recebimentos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% dos fluxos de autenticação ocorrem via conta Google, com zero senhas, hashes de senha ou tokens de recuperação de senha armazenados na base de dados.
- **SC-002**: 100% dos usuários previamente cadastrados conseguem acessar suas contas existentes ao entrar com o mesmo e-mail via Google no primeiro acesso, mantendo todo o seu histórico de ebooks, arquivos, clientes e vendas intacto.
- **SC-003**: Novos usuários conseguem realizar o login inicial via Google em menos de 10 segundos e concluir o onboarding complementar de criador em menos de 2 minutos.
- **SC-004**: 100% das rotas da aplicação (tanto autenticadas quanto públicas) respondem exclusivamente por meio de contratos de API estruturados (ou fluxos binários/CSV quando aplicável, como download de PDF e exportação CSV), com zero respostas dependendo de renderização de templates HTML no servidor (SSR).
- **SC-005**: Redução de 100% nos chamados e envios de e-mail transacional referentes a redefinição de senha e confirmação manual de e-mail no cadastro.

## Assumptions

- Clientes consumidores da API (como uma aplicação web SPA separada) serão responsáveis por toda a camada de interface gráfica e navegação do usuário.
- Apenas contas Google com endereço de e-mail confirmado (`email_verified = true`) são aceitas para login e criação de conta.
- Endpoints que entregam arquivos por natureza de negócio (download do PDF com marca d'água pelo comprador e exportação CSV de clientes pelo criador) continuam entregando seus respectivos fluxos de arquivo (PDF/CSV) em caso de sucesso, mas retornam erros estruturados em JSON quando houver falha de validação, autenticação ou permissão.
