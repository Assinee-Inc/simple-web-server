# Componente de Paginação Unificado

## Descrição
O sistema agora possui um componente de paginação unificado localizado em `/web/partials/transaction-table-footer.html` que pode ser usado em todas as páginas que necessitam de paginação.

## Templates Disponíveis

### `pagination-footer`
Template principal e recomendado para novos usos. Suporta múltiplos tipos de filtros.

### `transaction-table-footer` 
Template de compatibilidade que redireciona para `pagination-footer`.

### `table-footer`
Template específico para páginas de arquivos e outras que precisam de seleção de page size. 
**Localização:** `/web/partials/table_footer.html`
**Uso específico:** Páginas de clientes, arquivos, ebooks
**Características:** Inclui dropdown para alterar quantidade de itens por página

## Como Usar

### 1. Para páginas com filtros simples (recomendado)
```html
{{ if .Pagination }}
{{ template "pagination-footer" . }}
{{ end }}
```

### 2. Para páginas que precisam de seleção de page size
```html
{{ if .Pagination }}
{{ template "table-footer" . }}
{{ end }}
```

### 2. No handler
```go
// Adicionar dados de contexto
h.templateRenderer.View(w, r, "template_name", map[string]interface{}{
    "Pagination":  pagination,        // Obrigatório
    "RecordType":  "vendas",           // Opcional - personaliza o texto (ex: "vendas", "transações", "clientes")
    // Novo formato dinâmico de filtros
    "Filters": map[string]interface{}{
        "client_name":  clientName,    // Qualquer parâmetro de filtro
        "client_email": clientEmail,
        "ebook_id":     ebookID,
        "search":       searchTerm,
        "status":       status,
        // ... outros filtros conforme necessário
    },
}, layout)
```

## Formato Dinâmico de Filtros

O componente agora usa um formato dinâmico para filtros através do parâmetro `Filters`. Isso permite:

- **Flexibilidade total**: Qualquer parâmetro pode ser adicionado
- **Manutenibilidade**: Não há filtros hardcoded no template
- **Extensibilidade**: Novos filtros podem ser adicionados sem modificar o template

### Exemplo de migração:
```go
// ANTES (filtros fixos)
"ClientName": clientName,
"Status": status,

// DEPOIS (filtros dinâmicos)
"Filters": map[string]interface{}{
    "client_name": clientName,
    "status": status,
},
```

## Parâmetros Suportados

| Parâmetro | Tipo | Obrigatório | Descrição |
|-----------|------|-------------|-----------|
| `Pagination` | `*models.Pagination` | Sim | Objeto de paginação |
| `RecordType` | `string` | Não | Tipo de registro (ex: "vendas", "transações") |
| `Filters` | `map[string]interface{}` | Não | Mapa dinâmico de filtros para manter na URL |

### Filtros Dinâmicos (dentro de `Filters`)

Qualquer chave/valor pode ser adicionado ao mapa `Filters`. Exemplos comuns:

| Filtro | Tipo | Descrição |
|--------|------|-----------|
| `search` | `string` | Termo de busca geral |
| `status` | `string` | Status para filtro |
| `client_name` | `string` | Nome do cliente para filtro |
| `client_email` | `string` | Email do cliente para filtro |
| `ebook_id` | `uint` | ID do ebook para filtro |
| `type` | `string` | Tipo de arquivo ou categoria |
| `date_from` | `string` | Data inicial |
| `date_to` | `string` | Data final |

## Exemplo de Uso Completo

### Handler
```go
pagination := models.NewPagination(page, limit)
pagination.SetTotal(total)

h.templateRenderer.View(w, r, "purchase/list", map[string]interface{}{
    "Items":      items,
    "Pagination": pagination,
    "RecordType": "vendas",
    "ClientName": clientName,
    "EbookID":    ebookID,
}, "admin")
```

### Template
```html
{{ if and .Items (gt (len .Items) 0) }}
<div class="table-responsive">
    <!-- Sua tabela aqui -->
</div>

{{ if .Pagination }}
{{ template "pagination-footer" . }}
{{ end }}
{{ else }}
<!-- Estado vazio -->
{{ end }}
```

## Páginas Atualmente Usando os Componentes

### Usando `pagination-footer` (unificado)
1. **Transações** (`/web/pages/transactions/list.html`)
   - Usa `transaction-table-footer` (compatibilidade → `pagination-footer`)
   - Parâmetros: `Search`, `Status`

2. **Vendas** (`/web/pages/purchase/list.html`)
   - Usa `pagination-footer` (direto)
   - Parâmetros: `ClientName`, `ClientEmail`, `EbookID`, `RecordType`

### Usando `table-footer` (específico)
1. **Clientes** (`/web/pages/client.html`)
   - Parâmetros: `SearchTerm`, `PageSize`

2. **Arquivos** (`/web/pages/file/index.html`)
   - Parâmetros: `SearchTerm`, `FileType`, `PageSize`

3. **Ebooks** (`/web/pages/ebook/index.html`, `/create.html`, `/update.html`)
   - Parâmetros: `SearchTerm`, `PageSize`

## Migração

Para migrar páginas existentes:

1. Substitua templates customizados por `{{ template "pagination-footer" . }}`
2. Adicione `RecordType` no handler para personalizar o texto
3. Certifique-se de que todos os parâmetros de filtro estão no contexto
4. Remova templates de paginação duplicados

## Benefícios

- **Consistência**: Aparência e comportamento uniformes
- **Manutenibilidade**: Alterações em um só lugar
- **Flexibilidade**: Suporta diferentes tipos de filtros
- **Compatibilidade**: Mantém funcionamento de código existente
