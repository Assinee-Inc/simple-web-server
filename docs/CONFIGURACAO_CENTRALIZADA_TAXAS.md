# Configuração Centralizada de Taxas de Negócio

## 📋 Visão Geral

Este documento descreve a implementação da configuração centralizada para taxas de negócio, especificamente a taxa da plataforma sobre vendas de ebooks.

## 🎯 Problema Resolvido

**Antes:** A taxa da plataforma (5%) estava espalhada em múltiplos locais do código, causando:
- Inconsistências entre Stripe e banco de dados
- Dificuldade de manutenção
- Risco de erros futuros

**Depois:** Taxa centralizada em configuração única, garantindo consistência.

## 🏗️ Arquitetura da Solução

### 1. **Configuração Centralizada**

#### `internal/config/business.go`
```go
type BusinessConfig struct {
    PlatformFeePercentage float64 // 5% para a plataforma
    // ... outras configurações
}

var Business = BusinessConfig{
    PlatformFeePercentage: 0.05, // 5% centralizado
}
```

#### `internal/config/config.go`
```go
type AppConfiguration struct {
    PlatformFeePercentage float64 // Carregada do ambiente
}
```

### 2. **Variável de Ambiente**

No arquivo `.env`:
```bash
# Taxa da plataforma sobre vendas (0.05 = 5%)
PLATFORM_FEE_PERCENTAGE=0.05
```

### 3. **Métodos Utilitários**

```go
// Calcula taxa da plataforma
func (bc *BusinessConfig) GetPlatformFeeAmount(totalAmount int64) int64

// Calcula valor do criador
func (bc *BusinessConfig) GetCreatorAmount(totalAmount int64) int64

// Calcula taxa do Stripe
func (bc *BusinessConfig) GetStripeProcessingFee(totalAmount int64) int64
```

## 📍 Locais Atualizados

### Handlers
- ✅ `internal/handler/checkout_handler.go` - Usa `config.Business.PlatformFeePercentage`
- ✅ `internal/handler/stripe.go` - Usa `config.Business.GetPlatformFeeAmount()`

### Services
- ✅ `internal/service/transaction_service.go` - Usa `config.Business.PlatformFeePercentage`

### Models
- ✅ `internal/models/transaction.go` - Usa `config.Business.PlatformFeePercentage`

## 🔧 Como Usar

### Para Desenvolvedores

1. **Acessar a taxa atual:**
   ```go
   fee := config.Business.PlatformFeePercentage // 0.05 (5%)
   ```

2. **Calcular taxa em centavos:**
   ```go
   feeAmount := config.Business.GetPlatformFeeAmount(totalAmountCents)
   ```

3. **Calcular valor do criador:**
   ```go
   creatorAmount := config.Business.GetCreatorAmount(totalAmountCents)
   ```

### Para Administradores

1. **Alterar taxa via variável de ambiente:**
   ```bash
   PLATFORM_FEE_PERCENTAGE=0.03  # 3%
   PLATFORM_FEE_PERCENTAGE=0.07  # 7%
   ```

2. **Alterar em tempo de execução:**
   ```go
   config.Business.PlatformFeePercentage = 0.06 // 6%
   ```

## ✅ Benefícios

1. **Consistência:** Uma única fonte de verdade
2. **Manutenibilidade:** Mudança em um local só
3. **Configurabilidade:** Pode ser alterada via ambiente
4. **Testabilidade:** Fácil de testar com diferentes valores
5. **Flexibilidade:** Métodos utilitários para diferentes cálculos

## 🚨 Importante

- Sempre usar `config.Business.PlatformFeePercentage` em vez de valores hardcoded
- Não usar `0.05`, `5%` ou similar diretamente no código
- Usar os métodos utilitários quando possível
- Manter sincronização entre Stripe e banco de dados

## 🧪 Testes

Os testes foram atualizados para usar a configuração centralizada:
- `internal/models/transaction_test.go` - Testa com 5% padrão
- Todos os testes passando ✅

## 📚 Próximos Passos

Para futuras melhorias, considere:
1. Interface de administração para alterar taxas
2. Histórico de mudanças de taxas
3. Taxas diferentes por tipo de produto
4. Taxas progressivas baseadas no volume
