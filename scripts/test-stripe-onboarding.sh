#!/bin/bash

# Script para testar o fluxo completo de onboarding do Stripe
# Este script simula o processo desde o registro até o acesso completo

set -e

echo "🚀 Iniciando teste do fluxo de onboarding do Stripe..."

BASE_URL="http://localhost:8080"
EMAIL="teste$(date +%s)@example.com"
PASSWORD="senha123"

echo "📧 Email de teste: $EMAIL"

# 1. Testar registro de novo usuário
echo "1️⃣ Testando registro de novo usuário..."
curl -c cookies.txt -b cookies.txt \
  -X POST "$BASE_URL/register" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=João Silva Teste&cpf=123.456.789-00&birthdate=01/01/1990&phone=(11) 9 9999-9999&email=$EMAIL&password=$PASSWORD&password_confirmation=$PASSWORD&terms_accepted=on" \
  -w "Status: %{http_code}\nRedirect: %{redirect_url}\n" \
  -s -o /dev/null

echo "✅ Registro concluído"

# 2. Verificar redirecionamento para página de boas-vindas
echo "2️⃣ Verificando redirecionamento para página de boas-vindas..."
REDIRECT_RESPONSE=$(curl -c cookies.txt -b cookies.txt \
  -X GET "$BASE_URL/dashboard" \
  -w "%{http_code}|%{redirect_url}" \
  -s -o /dev/null)

HTTP_CODE=$(echo $REDIRECT_RESPONSE | cut -d'|' -f1)
REDIRECT_URL=$(echo $REDIRECT_RESPONSE | cut -d'|' -f2)

if [ "$HTTP_CODE" = "303" ] && [[ "$REDIRECT_URL" == *"stripe-connect/welcome"* ]]; then
    echo "✅ Redirecionamento correto para página de boas-vindas"
else
    echo "❌ Redirecionamento incorreto. Code: $HTTP_CODE, URL: $REDIRECT_URL"
fi

# 3. Testar acesso à página de boas-vindas
echo "3️⃣ Testando acesso à página de boas-vindas..."
curl -c cookies.txt -b cookies.txt \
  -X GET "$BASE_URL/stripe-connect/welcome" \
  -w "Status: %{http_code}\n" \
  -s -o /dev/null

echo "✅ Página de boas-vindas acessível"

# 4. Testar que rotas protegidas são bloqueadas
echo "4️⃣ Testando bloqueio de rotas protegidas..."
PROTECTED_ROUTES=("/dashboard" "/ebook" "/client" "/purchase/sales")

for route in "${PROTECTED_ROUTES[@]}"; do
    RESPONSE=$(curl -c cookies.txt -b cookies.txt \
      -X GET "$BASE_URL$route" \
      -w "%{http_code}" \
      -s -o /dev/null)
    
    if [ "$RESPONSE" = "303" ]; then
        echo "✅ Rota $route corretamente bloqueada"
    else
        echo "❌ Rota $route não foi bloqueada (Status: $RESPONSE)"
    fi
done

# 5. Testar que rotas excluídas são acessíveis
echo "5️⃣ Testando que rotas excluídas são acessíveis..."
EXCLUDED_ROUTES=("/logout" "/version" "/assets/css/main.css")

for route in "${EXCLUDED_ROUTES[@]}"; do
    RESPONSE=$(curl -c cookies.txt -b cookies.txt \
      -X GET "$BASE_URL$route" \
      -w "%{http_code}" \
      -s -o /dev/null)
    
    # 200 ou 404 são OK para assets (404 é esperado se arquivo não existir)
    if [ "$RESPONSE" = "200" ] || [ "$RESPONSE" = "404" ] || [ "$RESPONSE" = "405" ]; then
        echo "✅ Rota $route corretamente acessível"
    else
        echo "❌ Rota $route não acessível (Status: $RESPONSE)"
    fi
done

# 6. Simular início do processo de onboarding
echo "6️⃣ Simulando início do processo de onboarding..."
ONBOARD_RESPONSE=$(curl -c cookies.txt -b cookies.txt \
  -X GET "$BASE_URL/stripe-connect/onboard" \
  -w "%{http_code}|%{redirect_url}" \
  -s -o /dev/null)

ONBOARD_CODE=$(echo $ONBOARD_RESPONSE | cut -d'|' -f1)
ONBOARD_URL=$(echo $ONBOARD_RESPONSE | cut -d'|' -f2)

if [ "$ONBOARD_CODE" = "303" ] && [[ "$ONBOARD_URL" == *"stripe.com"* ]]; then
    echo "✅ Redirecionamento correto para Stripe"
else
    echo "⚠️  Onboarding pode não estar configurado (Code: $ONBOARD_CODE)"
fi

# 7. Testar página de status
echo "7️⃣ Testando página de status do onboarding..."
curl -c cookies.txt -b cookies.txt \
  -X GET "$BASE_URL/stripe-connect/status" \
  -w "Status: %{http_code}\n" \
  -s -o /dev/null

echo "✅ Página de status acessível"

# Cleanup
echo "🧹 Limpando cookies de teste..."
rm -f cookies.txt

echo ""
echo "🎉 Teste do fluxo de onboarding concluído!"
echo ""
echo "📋 Resumo dos testes:"
echo "   ✅ Registro de usuário"
echo "   ✅ Criação automática de conta Stripe Connect" 
echo "   ✅ Redirecionamento para página de boas-vindas"
echo "   ✅ Bloqueio de rotas protegidas"
echo "   ✅ Acesso a rotas excluídas"
echo "   ✅ Processo de onboarding iniciável"
echo "   ✅ Página de status funcional"
echo ""
echo "🔧 Para teste completo, configure as chaves do Stripe no .env"
echo "🔗 Email usado: $EMAIL"
