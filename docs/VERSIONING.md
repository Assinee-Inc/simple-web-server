# Versionamento do Projeto

Este documento descreve como o sistema de versionamento funciona no Docffy.

## Visão Geral

O projeto utiliza **versionamento semântico** (SemVer) seguindo o padrão `MAJOR.MINOR.PATCH`:
- **MAJOR**: Mudanças incompatíveis com versões anteriores
- **MINOR**: Novas funcionalidades compatíveis com versões anteriores
- **PATCH**: Correções de bugs compatíveis com versões anteriores

## Estrutura de Versionamento

### Arquivos de Configuração

- `internal/config/version.go` - Configuração de versão e informações de build
- `internal/handler/version_handler.go` - Handler para exibir informações de versão
- `Makefile` - Comandos de build e versionamento
- `Dockerfile` - Containerização com versionamento
- `scripts/release.sh` - Script de automação de releases

### Variáveis de Build

As seguintes variáveis são injetadas durante o build:

- `Version` - Versão atual (ex: v1.0.0)
- `CommitHash` - Hash do commit Git
- `BuildTime` - Timestamp do build
- `GoVersion` - Versão do Go utilizada

## Comandos Disponíveis

### Build

```bash
# Build local
make build

# Build para produção (Linux)
make build-prod

# Build para macOS
make build-mac

# Build para Windows
make build-windows

# Build para todas as plataformas
make build-all
```

### Versionamento

```bash
# Ver informações de versão
make version

# Criar tag de versão
make tag VERSION=v1.0.0

# Limpar artefatos de build
make clean
```

### Docker

```bash
# Build da imagem Docker
make docker-build

# Executar container
make docker-run
```

## Processo de Release

### 1. Release Automatizado

Use o script de release para automatizar o processo:

```bash
# Release patch (1.0.0 -> 1.0.1)
./scripts/release.sh patch

# Release minor (1.0.0 -> 1.1.0)
./scripts/release.sh minor

# Release major (1.0.0 -> 2.0.0)
./scripts/release.sh major
```

O script irá:
1. Verificar se há mudanças não commitadas
2. Executar testes
3. Fazer build da aplicação
4. Criar tag Git
5. Fazer push da tag
6. Build para múltiplas plataformas
7. Gerar notas de release

### 2. Release Manual

```bash
# 1. Atualizar versão no código (se necessário)
# 2. Commit das mudanças
git add .
git commit -m "feat: prepare for v1.0.0"

# 3. Criar tag
git tag -a v1.0.0 -m "Release v1.0.0"

# 4. Push da tag
git push origin v1.0.0

# 5. Build
make build-all
```

## Endpoints de Versão

A aplicação expõe endpoints para verificar a versão:

### Texto Simples
```bash
curl http://localhost:8080/version
```

Resposta:
```
Docffy
Version: v1.0.0
Commit: abc123def456
Build Time: 2024-01-15_10:30:00
Go Version: go1.23.4
```

### JSON
```bash
curl http://localhost:8080/api/version
```

Resposta:
```json
{
  "status": "success",
  "data": {
    "version": "v1.0.0",
    "commit_hash": "abc123def456",
    "build_time": "2024-01-15_10:30:00",
    "go_version": "go1.23.4"
  },
  "message": "Version information retrieved successfully"
}
```

## Docker com Versionamento

### Build com Argumentos

```bash
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg COMMIT_HASH=$(git rev-parse HEAD) \
  --build-arg BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S') \
  -t docffy:v1.0.0 .
```

### Build via Makefile

```bash
make docker-build VERSION=v1.0.0
```

## Convenções de Commit

Siga o padrão **Conventional Commits**:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Tipos de Commit

- `feat`: Nova funcionalidade
- `fix`: Correção de bug
- `docs`: Documentação
- `style`: Formatação, ponto e vírgula, etc.
- `refactor`: Refatoração de código
- `test`: Adicionar ou corrigir testes
- `chore`: Mudanças em build, ferramentas, etc.

### Exemplos

```bash
git commit -m "feat: add user authentication system"
git commit -m "fix: resolve database connection issue"
git commit -m "docs: update API documentation"
git commit -m "refactor: improve error handling in handlers"
```

## Estratégia de Branching

### Branches Principais

- `main` - Código de produção
- `develop` - Código em desenvolvimento
- `feature/*` - Novas funcionalidades
- `hotfix/*` - Correções urgentes
- `release/*` - Preparação de releases

### Workflow

1. **Feature Development**
   ```bash
   git checkout -b feature/nova-funcionalidade
   # Desenvolver funcionalidade
   git commit -m "feat: implement nova funcionalidade"
   git push origin feature/nova-funcionalidade
   # Criar Pull Request para develop
   ```

2. **Release Preparation**
   ```bash
   git checkout -b release/v1.0.0
   # Ajustes finais, atualizar versão
   git commit -m "chore: bump version to v1.0.0"
   # Merge para main e develop
   ```

3. **Hotfix**
   ```bash
   git checkout -b hotfix/critical-bug
   # Correção urgente
   git commit -m "fix: resolve critical security issue"
   # Merge para main e develop
   ```

## Monitoramento de Versão

### Health Check

O Dockerfile inclui health check que verifica o endpoint de versão:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/version || exit 1
```

### Logs

A versão é exibida nos logs de inicialização:

```
[INFO] Starting Docffy v1.0.0 (abc123def456)
[INFO] Build time: 2024-01-15_10:30:00
[INFO] Go version: go1.23.4
```

## Troubleshooting

### Problemas Comuns

1. **Erro de build com ldflags**
   - Verifique se o caminho do módulo está correto
   - Confirme se as variáveis estão definidas

2. **Tag não encontrada**
   - Execute `git fetch --tags` para buscar tags remotas
   - Verifique se a tag foi criada localmente

3. **Docker build falha**
   - Verifique se o Dockerfile está atualizado
   - Confirme se os argumentos de build estão corretos

### Debug

Para debug de versionamento:

```bash
# Ver variáveis de build
make version

# Testar endpoint de versão
curl http://localhost:8080/version

# Verificar tags Git
git tag -l

# Ver informações do último commit
git log --oneline -1
```

## Próximos Passos

1. Configurar CI/CD para builds automáticos
2. Implementar notificações de release
3. Adicionar métricas de versão
4. Criar dashboard de releases 