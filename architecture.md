# Diretrizes de Engenharia e Padrões de Projeto

AVISO CRÍTICO PARA LLM/IDE: Este projeto segue uma arquitetura rigorosa de Domínios Encapsulados. É estritamente proibido utilizar padrões legados ou pacotes fora da estrutura internal/. Ignore implementações antigas e siga este guia para qualquer criação ou manutenção de código.

1. Estrutura de Diretórios (Layout Obrigatório)
A organização deve ser orientada por funcionalidade (domínio), protegendo a lógica de negócio de detalhes de implementação técnica.

Hierarquia:

cmd/api/main.go: Ponto de entrada: Injeção de Dependências e Rotas.

internal/[domínio]/: Contém [domínio].go (interfaces), handler_api.go, handler_web.go, manager.go e repository.go.

internal/llm/: Integrações com IA (Gemini/OpenAI) via Interfaces.

internal/mailer/: Serviços de e-mail via Interfaces.

internal/platform/: Ferramentas globais (Validator, UUID, Response).

templates/: Arquivos .html organizados por domínio (SSR).

public/: Arquivos estáticos (CSS, JS, Imagens).

.
├── cmd/api/main.go          # Único ponto de entrada e injeção de dependências
├── internal/
│   ├── [domínio]/           # Ex: account, ebook, library
│   │   ├── [domínio].go     # Entidades puras e Interfaces de contrato
│   │   ├── handler_api.go   # Handlers para retorno JSON
│   │   ├── handler_web.go   # Handlers para SSR (Templates HTML)
│   │   ├── manager.go       # Lógica de Negócio (Obrigatório)
│   │   └── repository.go    # Persistência (Exclusivo GORM/SQL)
│   ├── llm/                 # Integrações com IA (Gemini/OpenAI) via Interfaces
│   ├── mailer/              # Serviços de e-mail via Interfaces
│   └── platform/            # Utilitários globais (Validator, UUID, Response)
├── templates/               # Arquivos .html organizados por [domínio]
└── public/                  # Ativos estáticos (CSS, JS)

2. Camadas e Responsabilidades
2.1 Handler (Entrega e Validação Sintática)

Papel: O "porteiro" da aplicação. Valida se o JSON ou o Formulário está correto.

Fail Fast: Utilize o validator/v10 com RegisterTagNameFunc para mapear erros para as chaves JSON.

Regra: O Manager só deve ser chamado se os dados estiverem sintaticamente perfeitos.

2.2 Manager (O Coração / Lógica de Negócio)

Papel: Onde residem as regras de negócio. É "cego" para o transporte (não sabe se é HTTP, JSON ou HTML).

Contratos: Deve consumir Repositories e Serviços Externos apenas através de Interfaces definidas no arquivo do domínio.

2.3 Repository (Persistência)

Papel: Implementação técnica do banco de dados (GORM). Não deve conter lógica de negócio.

3. Serviços Externos (Inversão de Dependência)
É PROIBIDO instanciar clientes de serviços externos diretamente dentro do Manager. Implementações técnicas residem em pacotes próprios (ex: internal/llm/gemini.go) e o Manager recebe a Interface no construtor.

4. Rigor de Testes (TEST-DRIVEN POLICY)
É obrigatório criar e manter testes para qualquer funcionalidade.

Handler Tests: Use Table-Driven Tests para validar falhas de validação sintática.

Manager Tests: Use testify/mock para simular dependências.

Isolamento: Separe o teste de "Sucesso" dos testes de falha em tabela.

5. Regras de Ouro para Código Idiomático
Contexto: O primeiro argumento de funções no Manager e Repository deve ser ctx context.Context.

Dinheiro: NUNCA use float64. Use int64 para representar centavos.

Erros: Utilize fmt.Errorf("contexto: %w", err) para empilhar erros.

ID Generation: Utilize a interface de platform/uuid para identificadores.

Nomes JSON: O validador deve sempre reportar erros usando o nome definido na tag json:.

6. Fluxo de Resposta (Padronização)
API: Respostas via ErrorResponse struct com campo fields para detalhamento.

SSR: Erros devem ser injetados no contexto do template para exibição no HTML.

Status Codes: 400 (JSON inválido), 422 (validação de campos), 500 (erro de servidor).