# Migração Visual para DaisyUI 4 + Tailwind CSS 3

## Contexto

Modernização visual do projeto migrando de Bootstrap (tema customizado `theme.min.css`) para **DaisyUI 4 + Tailwind CSS 3**, com migração gradual página a página.

### Por que DaisyUI 4 (e não 5)?

DaisyUI 5 requer Tailwind CSS 4 como plugin de build (`@plugin "daisyui"`). O browser CDN do Tailwind 4 não suporta plugins, tornando inviável o uso de DaisyUI 5 via CDN puro. DaisyUI 4 + Tailwind CDN v3 é a abordagem oficial documentada para uso sem build step. A migração para DaisyUI 5 pode ser feita futuramente ao adicionar um build step (ex: `bun` + CLI do Tailwind 4).

### Stack de CDN utilizada

```html
<!-- Componentes DaisyUI -->
<link href="https://cdn.jsdelivr.net/npm/daisyui@4/dist/full.min.css" rel="stylesheet" />
<!-- Utilitários Tailwind (flex, grid, responsivo, spacing) -->
<script src="https://cdn.tailwindcss.com"></script>
<!-- Ícones FontAwesome 6 (mantido do projeto original) -->
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.0/css/all.min.css" crossorigin="anonymous" />
```

---

## Estratégia de Migração Gradual

O template engine em `pkg/template/template.go` carrega todos os layouts de `web/layouts/*.html` e executa pelo nome. Isso permite criar novos layouts sem afetar páginas ainda não migradas.

### Regras durante a migração

1. **JS inline → arquivo externo**: Scripts encontrados no HTML devem ser movidos para `web/assets/js/<modulo>.<pagina>.js` e referenciados via `<script src="...">`.
2. **Bootstrap removido por layout**: Após migrar todas as páginas de um layout, remover Bootstrap daquele layout.
3. **Não quebrar páginas não migradas**: Layouts `guest.html` e `admin.html` permanecem com Bootstrap até sua fase de migração.

---

## Fases de Migração

### Fase 1 — Landing page ✅ CONCLUÍDO

| Arquivo | Status |
|---------|--------|
| `web/layouts/landing.html` | ✅ Criado (DaisyUI 4, sem Bootstrap) |
| `web/pages/home.html` | ✅ Migrado |
| `internal/shared/handler/home.go` | ✅ Atualizado (`"guest"` → `"landing"`) |
| JS inline extraído | N/A (home.html não tinha JS inline) |

**Detalhes técnicos:**
- Novo layout `web/layouts/landing.html` com DaisyUI 4 + Tailwind CDN v3
- Navbar responsiva com hamburger mobile usando CSS puro (DaisyUI `dropdown`)
- CSP atualizada em `pkg/middleware/security.go` para permitir `cdn.jsdelivr.net` e `cdn.tailwindcss.com`

---

### Fase 2 — Páginas de autenticação ✅ CONCLUÍDO

**Scope:** Migrar `guest.html` para DaisyUI. Bootstrap removido do `guest.html`.

| Arquivo | Status |
|---------|--------|
| `web/layouts/guest.html` | ✅ Migrado (DaisyUI 4, sem Bootstrap) |
| `web/pages/auth/login.html` | ✅ Migrado |
| `web/pages/auth/forget-password.html` | ✅ Migrado |
| `web/pages/auth/reset-password.html` | ✅ Migrado |
| `web/pages/auth/password-reset-success.html` | ✅ Migrado |
| `web/pages/creator/register.html` | ✅ Migrado |
| `web/assets/js/auth.login.js` | ✅ JS extraído do inline |
| `web/partials/form-errors-daisy.html` | ✅ Criado (DaisyUI alert, sem Bootstrap JS) |
| `web/partials/notifications-daisy.html` | ✅ Criado (DaisyUI alerts, sem Bootstrap Toast JS) |
| `web/partials/logo.html` | ✅ Atualizado para inline styles (sem dependência de framework) |

**Detalhes técnicos:**
- `guest.html` agora usa DaisyUI 4 + Tailwind CDN v3, sem Bootstrap
- Body usa `bg-base-200 min-h-screen flex items-center justify-center` para centrar cards
- `form-errors-daisy.html` substituiu `data-bs-dismiss` por alert DaisyUI puro (sem JS Bootstrap)
- `notifications-daisy.html` substituiu Bootstrap Toast por alerts DaisyUI inline (Success/Error/Info/Warning)
- `logo.html` migrado para `style=""` inline (funciona em qualquer contexto de layout)
- Toggle de senha em `login.html` extraído para `web/assets/js/auth.login.js` usando FontAwesome (`fa-eye`/`fa-eye-slash`)
- Classes jQuery Mask (`cpf`, `date`, `phone_with_ddd`) mantidas em `register.html` com `maxlength` como fallback; funcionalidade de máscara pendente de migração do `masks.js` (Fase 4)

---

### Fase 3 — Páginas públicas de produto ✅ CONCLUÍDO

| Arquivo | Status |
|---------|--------|
| `web/pages/creator/register.html` | ✅ Migrado (Fase 2) |
| `web/pages/purchase/checkout.html` | ✅ Migrado |
| `web/pages/purchase/sales_page.html` | ✅ Migrado |
| `web/pages/purchase/purchase-success.html` | ✅ Migrado |
| `web/pages/ebook/download.html` | ✅ Migrado |
| `web/pages/ebook/download-expired.html` | ✅ Migrado |
| `web/pages/ebook/download-limit-exceeded.html` | ✅ Migrado |
| `web/assets/js/purchase.checkout.js` | ✅ JS extraído (jQuery → vanilla JS + fetch) |
| `web/assets/js/purchase.sales-page.js` | ✅ JS extraído do inline |

**Detalhes técnicos:**
- `sales_page.html` tinha `<!DOCTYPE html>/<html>/<head>/<body>` embutidos dentro de `{{define "content"}}` (HTML malformado). Removidos — a página agora depende corretamente do layout `guest.html`
- `purchase-success.html` tinha tags `</head><body>` soltas no conteúdo. Corrigido
- `checkout.html` convertido de jQuery (`$.ajax`) para vanilla JS (`fetch`) em `purchase.checkout.js`
- Páginas `ebook/download*` são autônomas (`ViewWithoutLayout`) — Bootstrap substituído por DaisyUI 4 + Tailwind CDN diretamente no `<head>`
- Bootstrap Icons (`bi-*`) substituídos por FontAwesome equivalentes em todas as páginas

---

### Fase 4 — Dashboard e área autenticada (em andamento)

**Scope:** Migrar `admin.html` para DaisyUI. Após esta fase, Bootstrap é removido do `admin.html`.

**Estratégia:** Novo layout `web/layouts/admin-daisy.html` criado (análogo ao `landing.html` da Fase 1). Páginas migradas individualmente; ao concluir todas, `admin.html` é substituído.

| Arquivo | Status |
|---------|--------|
| `web/layouts/admin-daisy.html` | ✅ Criado (DaisyUI 4, drawer sidebar, sem Bootstrap) |
| `web/pages/dashboard.html` | ✅ Migrado |
| `web/assets/js/dashboard.js` | ✅ JS extraído do inline |
| `web/pages/settings.html` | ⬜ Migrar |
| `web/pages/billing.html` | ⬜ Migrar |
| `web/pages/ebook/index.html` | ⬜ Migrar |
| `web/pages/ebook/create.html` | ⬜ Migrar |
| `web/pages/ebook/update.html` | ⬜ Migrar |
| `web/pages/ebook/view.html` | ⬜ Migrar |
| `web/pages/ebook/send.html` | ⬜ Migrar |
| `web/pages/file/index.html` | ⬜ Migrar |
| `web/pages/file/upload.html` | ⬜ Migrar |
| `web/pages/client/list.html` | ⬜ Migrar |
| `web/pages/client/create.html` | ⬜ Migrar |
| `web/pages/client/update.html` | ⬜ Migrar |
| `web/pages/purchase/list.html` | ⬜ Migrar |
| `web/pages/transactions/list.html` | ⬜ Migrar |
| `web/pages/transactions/detail.html` | ⬜ Migrar |
| `web/pages/stripe-connect/*.html` | ⬜ Migrar |

**JS a extrair nesta fase:**
```
web/assets/js/ebook.view.js     → já existe (manter)
web/assets/js/file/file-preview.js → já existe (manter)
web/assets/js/masks.js          → já existe (avaliar migração de jQuery Mask)
```

---

### Fase 5 — Partials, error pages e templates (pendente)

| Arquivo | Status |
|---------|--------|
| `web/partials/*.html` | ⬜ Migrar todos |
| `web/pages/error/*.html` | ⬜ Migrar |
| `web/templates/base_page_template.html` | ⬜ Atualizar |
| `web/templates/list_page_template.html` | ⬜ Atualizar |

Após esta fase: remover diretório `web/assets/libs/bootstrap/` e demais libs não utilizadas.

---

## Mapeamento Bootstrap → DaisyUI (referência)

### Layout e Grid

| Bootstrap | DaisyUI / Tailwind |
|-----------|-------------------|
| `container` | `container mx-auto max-w-7xl px-4` |
| `row` | `grid grid-cols-1 md:grid-cols-N gap-N` ou `flex flex-wrap` |
| `col-lg-6` | `lg:w-1/2 w-full` |
| `col-lg-4` | `lg:w-1/3 w-full` |
| `d-flex align-items-center` | `flex items-center` |
| `justify-content-center` | `justify-center` |
| `d-none` / `d-lg-flex` | `hidden` / `lg:flex` |
| `ms-auto` | `ml-auto` |
| `me-N` | `mr-N` |
| `mb-N` / `mt-N` / `py-N` | Tailwind spacing equivalente |
| `g-4` (gap) | `gap-4` |

### Tipografia

| Bootstrap | Tailwind |
|-----------|---------|
| `display-1` | `text-7xl font-extrabold` |
| `display-3` | `text-5xl font-bold` |
| `display-4` | `text-4xl font-bold` |
| `lead` | `text-lg` |
| `fw-bold` | `font-bold` |
| `fw-semibold` | `font-semibold` |
| `text-uppercase` | `uppercase` |
| `ls-xl` (tracking) | `tracking-widest` |
| `text-decoration-none` | `no-underline` |
| `text-decoration-line-through` | `line-through` |
| `fs-3` | `text-3xl` |

### Componentes

| Bootstrap | DaisyUI |
|-----------|---------|
| `btn btn-primary` | `btn btn-primary` |
| `btn btn-outline-primary` | `btn btn-outline btn-primary` |
| `btn btn-light` | `btn bg-base-100 text-primary hover:bg-base-200` |
| `btn btn-outline-light` | `btn btn-outline text-primary-content` |
| `card` / `card-body` | `card bg-base-100 shadow-md` / `card-body` |
| `badge bg-primary` | `badge badge-primary` |
| `navbar navbar-expand-lg` | `navbar` + DaisyUI structure |
| `navbar-toggler` | `dropdown dropdown-end lg:hidden` |
| `dropdown-menu` | `dropdown-content` |
| `modal` (Bootstrap JS) | `modal` (DaisyUI CSS) |
| `toast` (Bootstrap JS) | `toast` + `alert` (DaisyUI CSS) |
| `form-control` | `input input-bordered` |
| `form-select` | `select select-bordered` |
| `form-check-input` | `checkbox` |
| `table` | `table` |

### Cores e Fundo

| Bootstrap | DaisyUI |
|-----------|---------|
| `bg-primary` | `bg-primary` |
| `bg-light` | `bg-base-200` |
| `bg-dark text-white` | `bg-neutral text-neutral-content` |
| `bg-white` | `bg-base-100` |
| `text-white` | `text-primary-content` ou `text-neutral-content` |
| `text-white-50` | `text-primary-content/75` |
| `text-muted` | `text-base-content/60` |
| `text-primary` | `text-primary` |
| `rounded-circle` | `rounded-full` |
| `rounded-3` / `rounded-4` | `rounded-xl` / `rounded-2xl` |
| `shadow-sm` / `shadow-lg` | `shadow-sm` / `shadow-xl` |
| `border-0` | `border-0` ou remover |

---

## Arquivos críticos

- `web/layouts/landing.html` — layout DaisyUI (páginas públicas migradas)
- `web/layouts/guest.html` — layout Bootstrap (auth, páginas públicas não migradas)
- `web/layouts/admin.html` — layout Bootstrap (área autenticada)
- `pkg/middleware/security.go` — CSP: adicionar CDNs necessários por fase
- `pkg/template/template.go` — engine de templates (não modificar)
