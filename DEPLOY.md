# Deploy no Heroku

Este documento contém as instruções para fazer deploy da aplicação Docffy no Heroku.

## Pré-requisitos

1. Conta no Heroku
2. Heroku CLI instalado
3. Git configurado

## Processo de Release e Deploy

### Fluxo Completo (Recomendado)

Para fazer um release e deploy completo:

```bash
# 1. Desenvolvimento e commits
git add .
git commit -m "feat: add new feature"

# 2. Release e deploy automático
./scripts/release-and-deploy.sh patch your-app-name
# ou
./scripts/release-and-deploy.sh minor your-app-name
# ou
./scripts/release-and-deploy.sh major your-app-name
```

### Fluxo Manual

Se preferir fazer o processo manualmente:

```bash
# 1. Criar release
./scripts/release.sh patch  # ou minor/major

# 2. Push para origin e Heroku
git push origin master
git push heroku master
```

## Tipos de Release

### Semantic Versioning

- **Patch** (`./scripts/release.sh patch`): Correções de bugs (1.0.0 → 1.0.1)
- **Minor** (`./scripts/release.sh minor`): Novas funcionalidades (1.0.0 → 1.1.0)
- **Major** (`./scripts/release.sh major`): Mudanças que quebram compatibilidade (1.0.0 → 2.0.0)

### Exemplos de Uso

```bash
# Correção de bug
./scripts/release-and-deploy.sh patch my-app

# Nova funcionalidade
./scripts/release-and-deploy.sh minor my-app

# Mudança significativa
./scripts/release-and-deploy.sh major my-app
```

## Configuração Inicial

### 1. Login no Heroku
```bash
heroku login
```

### 2. Criar aplicação no Heroku
```bash
heroku create your-app-name
```

### 3. Configurar buildpack
```bash
heroku buildpacks:set heroku/go
```

### 4. Configurar variáveis de ambiente
```bash
# Configurações básicas
heroku config:set APPLICATION_MODE=production
heroku config:set APPLICATION_NAME="Docffy"
heroku config:set APP_KEY=$(openssl rand -base64 32)

# Configurações de banco de dados (PostgreSQL)
heroku config:set DATABASE_URL=$(heroku config:get DATABASE_URL)

# Configurações de e-mail
heroku config:set MAIL_HOST=smtp.gmail.com
heroku config:set MAIL_PORT=587
heroku config:set MAIL_USERNAME=your-email@gmail.com
heroku config:set MAIL_PASSWORD=your-app-password
heroku config:set MAIL_FROM_ADDRESS=your-email@gmail.com

# Configurações do Stripe
heroku config:set STRIPE_SECRET_KEY=sk_test_...
heroku config:set STRIPE_PRICE_ID=price_...
heroku config:set STRIPE_WEBHOOK_SECRET=whsec_...

# Configurações do S3
heroku config:set S3_ACCESS_KEY=your-access-key
heroku config:set S3_SECRET_KEY=your-secret-key
heroku config:set S3_REGION=sa-east-1
heroku config:set S3_BUCKET_NAME=your-bucket-name

# Configurações da Receita Federal
heroku config:set HUB_DEVSENVOLVEDOR_API=your-api-url
heroku config:set HUB_DEVSENVOLVEDOR_TOKEN=your-token
```

### 5. Adicionar addon do PostgreSQL
```bash
heroku addons:create heroku-postgresql:mini
```

## Deploy

### 1. Commit das alterações
```bash
git add .
git commit -m "feat: prepare for heroku deployment"
```

### 2. Deploy
```bash
git push heroku master
```

### 4. Verificar logs
```bash
heroku logs --tail
```

### 5. Abrir aplicação
```bash
heroku open
```

## Estrutura de Arquivos para Deploy

### Arquivos Necessários:
- `Procfile` - Define o comando de execução
- `heroku.yml` - Configuração de build
- `app.json` - Metadados da aplicação
- `go.mod` - Dependências Go
- `go.sum` - Checksums das dependências
- `cmd/web/main.go` - Ponto de entrada da aplicação
- `web/` - Assets estáticos
- `internal/` - Código da aplicação
- `pkg/` - Pacotes utilitários

### Configurações Importantes:

1. **Porta**: A aplicação usa a variável `PORT` que o Heroku define automaticamente
2. **Banco de dados**: Use PostgreSQL em produção (SQLite apenas para desenvolvimento)
3. **Assets estáticos**: Servidos diretamente pela aplicação Go
4. **Rate limiting**: Configurado para produção

## Troubleshooting

### Problemas Comuns:

1. **Build falha**: Verifique se todas as dependências estão no `go.mod`
2. **Aplicação não inicia**: Verifique os logs com `heroku logs --tail`
3. **Banco de dados não conecta**: Verifique se o addon PostgreSQL está ativo
4. **Assets não carregam**: Verifique se a pasta `web/` está incluída no deploy

### Comandos Úteis:

```bash
# Ver configurações
heroku config

# Ver logs em tempo real
heroku logs --tail

# Executar comando na aplicação
heroku run bash

# Reiniciar aplicação
heroku restart

# Ver status da aplicação
heroku ps
```

## Monitoramento

### Logs
```bash
# Ver logs recentes
heroku logs

# Ver logs em tempo real
heroku logs --tail

# Ver logs de uma hora específica
heroku logs --since 1h
```

### Métricas
```bash
# Ver uso de recursos
heroku ps

# Ver addons
heroku addons
```

## Segurança

### Variáveis Sensíveis
- Nunca commite senhas ou chaves no código
- Use `heroku config:set` para variáveis sensíveis
- Use `heroku config:get` para verificar configurações

### SSL
- O Heroku fornece SSL automaticamente
- Configure `HOST` para usar `https://`

## Escalabilidade

### Dynos
- **Basic**: Para desenvolvimento e testes
- **Standard**: Para produção com tráfego moderado
- **Performance**: Para alta demanda

### Banco de Dados
- **Mini**: Para desenvolvimento
- **Basic**: Para produção
- **Standard**: Para alta demanda

## Backup

### Banco de Dados
```bash
# Backup manual
heroku pg:backups:capture

# Download do backup
heroku pg:backups:download

# Restaurar backup
heroku pg:backups:restore b001 DATABASE_URL
```

## Versionamento e Tags

### Estrutura de Tags

As tags seguem o padrão Semantic Versioning:
- `v1.0.0` - Primeira versão estável
- `v1.0.1` - Correção de bug
- `v1.1.0` - Nova funcionalidade
- `v2.0.0` - Mudança significativa

### Comandos de Versionamento

```bash
# Ver versão atual
git describe --tags --abbrev=0

# Ver histórico de tags
git tag -l

# Ver commits desde última tag
git log --oneline $(git describe --tags --abbrev=0)..HEAD

# Criar tag manualmente
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1
```

### Integração com CI/CD

Para integração com GitHub Actions ou outros CI/CD:

```yaml
# Exemplo de workflow GitHub Actions
name: Release and Deploy
on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy to Heroku
        run: |
          git push heroku ${{ github.ref_name }}
``` 