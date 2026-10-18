# Regras do Projeto Docffy

Este diretório contém todas as regras que devem ser seguidas durante o desenvolvimento do projeto.

## Regras Disponíveis

### 1. **analysis-rule.mdc** ⭐ **OBRIGATÓRIA**
- **Objetivo**: Sempre fazer análise completa antes de modificações
- **Quando aplicar**: ANTES de qualquer modificação no código
- **Prioridade**: MÁXIMA

### 2. **handler-rule.mdc**
- **Objetivo**: Padrões para criação de handlers HTTP
- **Quando aplicar**: Ao criar ou modificar handlers
- **Localização**: `internal/handler/`

### 3. **service-rule.mdc**
- **Objetivo**: Padrões para criação de serviços
- **Quando aplicar**: Ao criar ou modificar serviços
- **Localização**: `internal/service/`

### 4. **repository-rule.mdc**
- **Objetivo**: Padrões para criação de repositórios
- **Quando aplicar**: Ao criar ou modificar repositórios
- **Localização**: `internal/repository/`

### 5. **html-rule.mdc**
- **Objetivo**: Padrões para templates HTML
- **Quando aplicar**: Ao criar ou modificar templates
- **Localização**: `web/`

### 6. **test-rule.mdc**
- **Objetivo**: Padrões para criação de testes
- **Quando aplicar**: Ao criar ou modificar testes
- **Localização**: `*_test.go`

### 7. **ui-design-rules.mdc**
- **Objetivo**: Padrões de design da interface
- **Quando aplicar**: Ao criar ou modificar componentes UI
- **Localização**: `web/`

### 8. **pagination-rules.mdc**
- **Objetivo**: Padrões para implementação de paginação
- **Quando aplicar**: Ao implementar listagens paginadas
- **Localização**: Handlers e templates

### 9. **archictecture-rules.mdc**
- **Objetivo**: Regras gerais de arquitetura
- **Quando aplicar**: Sempre
- **Localização**: Todo o projeto

## Ordem de Aplicação

1. **SEMPRE** começar com `analysis-rule.mdc`
2. Aplicar regras específicas do tipo de modificação
3. Verificar regras de arquitetura
4. Aplicar regras de UI se aplicável

## Checklist Obrigatório

Antes de qualquer modificação:

- [ ] Análise completa do problema (analysis-rule)
- [ ] Identificação de todos os componentes afetados
- [ ] Verificação das regras específicas aplicáveis
- [ ] Proposta de solução estruturada
- [ ] Confirmação do usuário antes da implementação

## Exemplo de Uso

```markdown
## Análise do Problema (analysis-rule.mdc)

### Problema Reportado
- Descrição do problema
- Erros específicos
- Impacto no usuário

### Componentes Afetados
- Lista de arquivos
- Dependências
- Regras aplicáveis

### Solução Proposta
1. Modificação 1
2. Modificação 2
3. Modificação 3

## Aplicação das Regras Específicas

### Handler (handler-rule.mdc)
- Verificar estrutura do handler
- Aplicar padrões de nomenclatura
- Implementar injeção de dependências

### Service (service-rule.mdc)
- Verificar interface e implementação
- Aplicar lógica de negócio
- Implementar validações

### Repository (repository-rule.mdc)
- Verificar acesso a dados
- Implementar queries otimizadas
- Aplicar padrões de CRUD
```

## Penalidades

- **NUNCA** ignorar a analysis-rule
- **NUNCA** fazer modificações sem análise prévia
- **NUNCA** quebrar padrões estabelecidos
- **NUNCA** implementar sem entender o contexto completo

## Suporte

Em caso de dúvidas sobre as regras:
1. Consultar a regra específica
2. Verificar exemplos no código existente
3. Seguir os padrões já estabelecidos
4. Em caso de conflito, priorizar a analysis-rule 