# Regras de Design UI - Docffy

## Visão Geral
Este documento define as regras de design UI para garantir consistência visual em todas as páginas do sistema Docffy.

## Estrutura Base das Páginas

### 1. Container Principal
```html
<div class="container-fluid p-6">
  <!-- Conteúdo da página -->
</div>
```

### 2. Header da Página
```html
<div class="row">
  <div class="col-lg-12 col-md-12 col-12">
    <div class="border-bottom pb-4 mb-4">
      <div class="row align-items-center">
        <div class="col">
          <h3 class="mb-0 fw-bold">Título da Página</h3>
          <p class="mb-0 text-muted">Descrição da página</p>
        </div>
        <div class="col-auto">
          <div class="d-flex gap-2">
            <!-- Botões de ação -->
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
```

### 3. Conteúdo Principal
```html
<div class="py-6">
  <div class="row">
    <div class="col-xl-12 col-lg-12 col-md-12 col-12">
      <div class="card h-100">
        <!-- Card body ou header conforme necessário -->
      </div>
    </div>
  </div>
</div>
```

## Tipos de Páginas

### 1. Páginas de Listagem (Index)
- **Estrutura**: Header + Card com tabela
- **Card Header**: Busca e filtros
- **Card Body**: Tabela responsiva
- **Exemplo**: `/ebook`, `/client`

### 2. Páginas de Formulário (Create/Update)
- **Estrutura**: Header + Card único com formulário
- **Layout**: 2 colunas (8/4) ou 1 coluna
- **Seções**: Blocos com títulos h5/h6, sem cards internos
- **Exemplo**: `/ebook/create`, `/ebook/update`

### 3. Páginas de Dashboard
- **Estrutura**: Background colorido + cards de estatísticas
- **Layout**: Grid responsivo de cards
- **Exemplo**: `/dashboard`

## Regras de Componentes

### 1. Títulos e Hierarquia
- **H3**: Título principal da página (fw-bold)
- **H5**: Seções principais em formulários (mb-3)
- **H6**: Seções secundárias na sidebar (mb-3)

### 2. Ícones
- **FontAwesome**: Usar sempre FontAwesome (fa-solid, fa-regular)
- **Tamanhos**: 
  - `icon-xs`: Para elementos pequenos
  - `icon-sm`: Para títulos e elementos médios
- **Posicionamento**: Sempre `me-2` após ícone

### 3. Botões
```html
<!-- Botão primário -->
<button class="btn btn-primary">
  <i class="fa-solid fa-plus icon-xs me-2"></i>
  Texto do Botão
</button>

<!-- Botão secundário -->
<button class="btn btn-outline-secondary">
  <i class="fa-solid fa-arrow-left icon-xs me-2"></i>
  Voltar
</button>
```

### 4. Formulários
```html
<div class="mb-3">
  <label for="field" class="form-label fw-semibold">
    Label do Campo <span class="text-danger">*</span>
  </label>
  <input type="text" class="form-control" id="field" name="field" required>
  <div class="form-text">
    <i class="fa-solid fa-lightbulb icon-xs me-1"></i>
    Dica ou informação adicional
  </div>
</div>
```

### 5. Alertas
```html
<div class="alert alert-info border-0">
  <div class="d-flex align-items-center">
    <i class="fa-solid fa-circle-info icon-sm me-2"></i>
    <div>
      <strong>Título:</strong> Mensagem do alerta
    </div>
  </div>
</div>
```

### 6. Tabelas
```html
<div class="table-responsive">
  <table class="table table-hover text-nowrap">
    <thead class="table-light">
      <tr>
        <th scope="col" class="border-0">Coluna</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td class="align-middle">
          <!-- Conteúdo da célula -->
        </td>
      </tr>
    </tbody>
  </table>
</div>
```

## Layout de Formulários

### 1. Estrutura de Seções
```html
<div class="col-lg-8">
  <div class="mb-4">
    <h5 class="mb-3">
      <i class="fa-solid fa-circle-info icon-sm me-2"></i>
      Título da Seção
    </h5>
    <!-- Conteúdo da seção -->
  </div>
</div>
```

### 2. Sidebar (Coluna Direita)
```html
<div class="col-lg-4">
  <div class="mb-4">
    <h6 class="mb-3">
      <i class="fa-regular fa-lightbulb icon-sm me-2"></i>
      Título da Seção
    </h6>
    <!-- Conteúdo da sidebar -->
  </div>
</div>
```

## Cores e Estilos

### 1. Cores Primárias
- **Primary**: Azul do Bootstrap
- **Success**: Verde para sucesso/ativo
- **Warning**: Amarelo para avisos
- **Danger**: Vermelho para erros/perigo
- **Secondary**: Cinza para elementos secundários

### 2. Backgrounds
- **Cards**: `bg-white` (padrão)
- **Headers**: `bg-white` ou `bg-transparent`
- **Tabelas**: `table-light` para cabeçalhos

### 3. Bordas e Sombras
- **Cards principais**: `shadow-sm` (quando necessário)
- **Cards internos**: Sem bordas/sombras (apenas espaçamento)
- **Bordas**: `border-0` para remover bordas padrão

## Responsividade

### 1. Breakpoints
- **xs**: < 576px
- **sm**: ≥ 576px
- **md**: ≥ 768px
- **lg**: ≥ 992px
- **xl**: ≥ 1200px

### 2. Classes Responsivas
- **Colunas**: `col-lg-8 col-md-12 col-12`
- **Visibilidade**: `d-none d-md-block`
- **Tamanhos**: `btn-sm`, `form-select-sm`

## Validação e Estados

### 1. Campos Obrigatórios
```html
<span class="text-danger">*</span>
```

### 2. Mensagens de Erro
```html
<div class="text-danger small mt-1">
  <i class="fa-solid fa-circle-exclamation icon-xs me-1"></i>
  Mensagem de erro
</div>
```

### 3. Estados de Loading
```html
<button class="btn btn-primary" disabled>
  <i class="fa-solid fa-spinner icon-sm me-1 animate-spin"></i>
  Carregando...
</button>
```

## Padrões de Nomenclatura

### 1. IDs e Names
- **IDs**: camelCase (ex: `createEbookForm`)
- **Names**: snake_case (ex: `selected_files`)
- **Classes**: kebab-case (ex: `btn-outline-primary`)

### 2. Variáveis de Template
- **Dados**: `.VariableName`
- **Formulário**: `.Form.FieldName`
- **Erros**: `.Errors.FieldName`

## JavaScript e Interatividade

### 1. Event Listeners
```javascript
document.addEventListener('DOMContentLoaded', function() {
  // Inicialização
});
```

### 2. Validação de Formulários
```javascript
document.getElementById('formId').addEventListener('submit', function(e) {
  // Validação
  if (!isValid) {
    e.preventDefault();
    return;
  }
  
  // Loading state
  const submitBtn = document.getElementById('submitBtn');
  submitBtn.disabled = true;
  submitBtn.innerHTML = '<i class="fa-solid fa-spinner icon-sm me-1 animate-spin"></i> Salvando...';
});
```

## Checklist para Novas Páginas

- [ ] Estrutura base com container-fluid
- [ ] Header com título e botões de ação
- [ ] Layout responsivo (col-lg-*)
- [ ] Ícones FontAwesome consistentes
- [ ] Classes de espaçamento (mb-*, p-*)
- [ ] Validação de formulários
- [ ] Estados de loading
- [ ] Mensagens de erro/feedback
- [ ] Responsividade testada

## Exceções e Variações

### 1. Dashboard
- Background colorido no topo
- Cards de estatísticas sem header
- Layout em grid

### 2. Páginas de Erro
- Layout simplificado
- Foco na mensagem de erro
- Botão de retorno

### 3. Páginas de Login/Auth
- Layout centralizado
- Formulário simples
- Sem sidebar

---

**Última atualização**: Dezembro 2024
**Versão**: 1.0 