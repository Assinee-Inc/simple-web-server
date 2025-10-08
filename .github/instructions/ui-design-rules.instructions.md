---
applyTo: "**/*.html, **/*.css, **/*.scss"
description: "Regras de design de UI e diretrizes visuais para componentes, paginação, formulários e responsividade"
---

# UI Design & Visual Guidelines

## Componentes de Estatísticas

### Cards de Estatísticas

- **Container**: Use `card border-0 shadow-sm bg-light` para fundo suave
- **Título**: Inclua ícone colorido em círculo com `avatar-sm rounded-circle bg-{color}`
- **Espaçamento**: Use `mb-3` entre cards

### Quadrados de Métricas

- **Estrutura**: `text-center p-3 rounded border bg-white`
- **Altura uniforme**: SEMPRE use `min-height: 80px` (criação) ou `min-height: 70px` (edição)
- **Centralização**: SEMPRE use `display: flex; flex-direction: column; justify-content: center;`
- **Layout**: Grid responsivo com `col-6` para 2 métricas ou `col-6` para 4 métricas

### Tipografia em Métricas

- **Números**: Use `h3` ou `h4` com `fw-bold` e cores semânticas
- **Labels**: Use `small` com `text-muted fw-medium`
- **Texto curto**: Máximo 8-10 caracteres para evitar quebras de linha

### Código de Cores para Métricas

- **Azul (Primary)**: Contadores principais, arquivos, itens
- **Verde (Success)**: Valores monetários, tamanhos, resultados positivos
- **Ciano (Info)**: Métricas de engajamento, visualizações, interações
- **Laranja (Warning)**: Métricas de conversão, vendas, alertas importantes
- **Cinza (Muted)**: Labels e textos secundários

## Paginação

### Estrutura de Paginação

- **Container**: `d-flex justify-content-between align-items-center mt-3`
- **Navegação**: Use `pagination pagination-sm mb-0`
- **Separadores**: Use `-1` no array para representar "..." com `disabled`

### Informações de Paginação

- **Formato**: "Página X de Y (Z por página)"
- **Classes**: `text-muted small` para informações contextuais
- **Total**: Sempre mostrar total de registros disponíveis

## Sidebar de Criação/Edição

### Dicas e Ajuda

- **Card**: `card bg-light mb-3` para fundo neutro
- **Ícones**: Use FontAwesome
- **Lista**: `list-unstyled mb-0` com check icons verdes

### Diferenciação de Páginas

- **Criação**: Ícones azuis (Primary), métricas em tempo real
- **Edição**: Ícones verdes (Success), métricas do ebook existente

## Formulários e Controles

### Botões de Ação

- **Primário**: `btn btn-primary` para ações principais
- **Secundário**: `btn btn-secondary` para cancelar/voltar
- **Layout**: Use `d-grid gap-2` para empilhamento vertical

### Seleção de Arquivos

- **Tabela**: `table table-hover` para listagem
- **Controles**: Botões "Selecionar Todos" e "Desmarcar Todos"
- **Busca**: Input com placeholder "Buscar arquivos..."
- **Estatísticas**: Contador em tempo real de seleções

## Responsive Design

### Breakpoints

- **Desktop**: Layout completo com sidebar
- **Mobile**: Stack vertical, sidebar abaixo do conteúdo principal
- **Grid**: Use classes Bootstrap responsivas (`col-md-8`, `col-md-4`)

### Espaçamentos

- **Entre seções**: `mb-4` ou `mb-3`
- **Elementos relacionados**: `g-2` ou `g-3` em grids
- **Padding interno**: `p-3` para cards, `p-2` para elementos menores

## Padrões de Cores

### Backgrounds

- **Cards principais**: `bg-white` com `border` sutil
- **Cards secundários**: `bg-light` para diferenciação
- **Hover states**: `table-hover` para interatividade

### Text Colors

- **Primário**: `text-dark` para títulos
- **Secundário**: `text-muted` para labels e descrições
- **Métricas**: Cores semânticas baseadas no contexto

## Ícones MDI

### Padrão de Uso

- **Estatísticas**: `mdi-chart-line`
- **Arquivos**: `mdi-file-*` baseado no tipo
- **Ações**: `mdi-plus`, `mdi-pencil`, `mdi-content-save`
- **Navegação**: `mdi-chevron-left`, `mdi-chevron-right`
- **Status**: `mdi-check-circle` para sucesso

### Posicionamento

- **Com texto**: Use `me-1` ou `me-2` para espaçamento
- **Em botões**: Ícone + espaço + texto
- **Em avatars**: Centralizados com flexbox

## Validações de Design

### Checklist de Qualidade

- [ ] Quadrados de métricas têm altura uniforme
- [ ] Textos não quebram linha desnecessariamente
- [ ] Cores seguem código semântico estabelecido
- [ ] Espaçamentos são consistentes
- [ ] Layout é responsivo em todas as telas
- [ ] Contraste é adequado para acessibilidade

### Teste Visual

- Verificar alinhamento em diferentes resoluções
- Testar com conteúdo real (textos longos/curtos)
- Validar hierarquia visual e escaneabilidade
- Confirmar que ícones estão alinhados com texto

## Exemplo de Implementação

```html
<!-- Card de Estatísticas Padrão -->
<div class="card border-0 shadow-sm mb-3 bg-light">
  <div class="card-body">
    <h6 class="card-title mb-3 d-flex align-items-center text-dark">
      <div
        class="avatar-sm rounded-circle bg-primary d-flex align-items-center justify-content-center me-2"
      >
        <i class="mdi mdi-chart-line text-white"></i>
      </div>
      Título da Seção
    </h6>
    <div class="row g-3">
      <div class="col-6">
        <div
          class="text-center p-3 rounded border bg-white"
          style="min-height: 80px; display: flex; flex-direction: column; justify-content: center;"
        >
          <h3 class="mb-1 text-primary fw-bold">0</h3>
          <small class="text-muted fw-medium">Métrica</small>
        </div>
      </div>
    </div>
  </div>
</div>
```
