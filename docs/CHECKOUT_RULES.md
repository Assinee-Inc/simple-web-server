# Regras de Negócio do Checkout — Docffy

**Produto:** Docffy
**Domínio:** Vendas / Checkout de E-books
**Versão:** 1.0
**Data:** 24/03/2026
**Responsável:** Lead Product Owner

---

## 1. Visão Geral

O fluxo de checkout da Docffy é o processo pelo qual um comprador adquire um e-book digital publicado por um criador de conteúdo cadastrado na plataforma. O objetivo do fluxo é coletar e validar os dados do comprador, processar o pagamento de forma segura via Stripe, registrar a transação e disponibilizar o acesso ao conteúdo adquirido — tudo de forma auditável e sem ambiguidade sobre o estado da compra.

O checkout é composto por três etapas principais: apresentação do produto, validação do comprador e processamento do pagamento. A confirmação da compra ocorre exclusivamente por meio de notificação assíncrona do Stripe (webhook), garantindo que nenhuma ação de negócio seja executada com base apenas na navegação do usuário.

---

## 2. Pré-condições

Para que um checkout possa ser iniciado, todas as condições abaixo devem ser satisfeitas simultaneamente:

| # | Condição | Comportamento quando não satisfeita |
|---|---|---|
| 2.1 | O e-book referenciado pela URL deve existir no sistema | Erro 404 — página de produto não encontrada |
| 2.2 | O e-book deve estar com status **publicado** | Erro 404 — produto indisponível para venda |
| 2.3 | O criador associado ao e-book deve estar cadastrado e ativo | Checkout bloqueado |
| 2.4 | O valor do e-book deve estar definido e ser maior que zero | Checkout bloqueado |

Um e-book não publicado nunca pode ser vendido, independentemente de como a URL for acessada.

---

## 3. Regras de Validação do Cliente

Antes de criar qualquer registro de compra, os dados do comprador são validados integralmente. A validação falha se qualquer campo estiver ausente, com formato inválido ou não passar nas verificações externas.

### 3.1 Campos obrigatórios e formatos aceitos

| Campo | Obrigatoriedade | Regra de formato |
|---|---|---|
| Nome | Obrigatório | Não pode ser vazio |
| CPF | Obrigatório | Exatamente 11 caracteres numéricos, sem pontuação |
| Data de nascimento | Obrigatório | Formato DD/MM/AAAA |
| E-mail | Obrigatório | Entre 3 e 254 caracteres, formato de endereço de e-mail válido |
| Telefone | Obrigatório | Exatamente 11 caracteres numéricos, sem formatação |
| Identificador do e-book | Obrigatório | Deve corresponder a um e-book existente e publicado |

Todos os campos são validados antes de qualquer consulta ao banco de dados ou serviço externo.

### 3.2 Validação de CPF via Receita Federal

Em ambiente de produção, o CPF informado é verificado junto à API da Receita Federal (via Hub Desenvolvedor Gov.br). As seguintes regras se aplicam:

- O CPF deve estar registrado e ativo na base da Receita Federal.
- O nome informado pelo comprador deve corresponder ao nome cadastrado no CPF consultado. Divergência bloqueia o checkout.
- Em ambiente de desenvolvimento, esta verificação externa é suprimida e apenas o formato do CPF é validado.

A validação do CPF via Receita Federal serve como camada de prevenção contra uso de identidades falsas ou inconsistentes.

---

## 4. Regras de Prevenção de Duplicidade e Fraude

### 4.1 Unicidade de CPF

O CPF é o identificador único de um comprador na plataforma. Não podem existir dois compradores cadastrados com o mesmo CPF. Em caso de reuso do mesmo CPF em uma nova tentativa de compra, o sistema atualiza o e-mail associado ao comprador existente — sem criar um novo registro.

### 4.2 Prevenção de compra duplicada

Um mesmo comprador (identificado pelo CPF) não pode adquirir o mesmo e-book mais de uma vez. Se a tentativa de compra for detectada como duplicata, o sistema:

- Bloqueia o prosseguimento do checkout;
- Retorna uma mensagem de erro informando que o comprador já possui o produto;
- Disponibiliza os dados de contato do criador para que o comprador possa solicitar suporte diretamente.

Esta regra protege tanto o comprador (contra cobranças indevidas) quanto o criador (contra chargebacks por duplicidade).

### 4.3 Validação de identidade em produção

Conforme descrito na seção 3.2, a verificação ativa do CPF junto à Receita Federal em produção constitui uma barreira adicional contra fraudes de identidade. Qualquer inconsistência entre os dados informados e os dados oficiais interrompe o fluxo imediatamente.

---

## 5. Regras de Criação de Compra

### 5.1 Momento de criação

O registro de compra é criado no sistema **antes** do redirecionamento para o ambiente de pagamento do Stripe. Isso garante rastreabilidade da intenção de compra independentemente do desfecho do pagamento.

### 5.2 Estado inicial

Toda compra recém-criada assume o estado inicial descrito abaixo:

| Atributo | Valor inicial | Significado |
|---|---|---|
| Status | Pendente | Pagamento ainda não confirmado |
| Limite de downloads | Ilimitado (-1) | Sem restrição por padrão |
| Expiração do acesso | Sem expiração (nulo) | Acesso permanente por padrão |
| Identificador de URL | UUID v7 gerado automaticamente | Usado na URL de download — nunca o ID interno |

### 5.3 Identificador público

O acesso ao conteúdo adquirido é sempre referenciado por um identificador público gerado por UUID v7 (campo `HashID`). O ID interno sequencial nunca é exposto em URLs ou comunicações externas.

### 5.4 Transação associada

Junto à compra, é criado um registro de transação financeira com status **pendente**. Este registro acompanha o ciclo de vida do pagamento e é atualizado quando o Stripe confirmar ou rejeitar a cobrança.

---

## 6. Regras de Pagamento

### 6.1 Processador de pagamento

Todo pagamento na Docffy é processado exclusivamente via **Stripe Checkout**, em sessão hospedada pelo Stripe. O comprador é redirecionado para o ambiente seguro do Stripe para inserir os dados de pagamento.

### 6.2 Moeda

Todos os valores são processados em **Real Brasileiro (BRL)**.

### 6.3 Tipo de sessão

O checkout utiliza o modo de pagamento único (`payment`), adequado para transações de compra avulsa sem recorrência.

### 6.4 Página de sucesso

Após o pagamento ser processado pelo Stripe, o comprador é redirecionado para uma página de confirmação na Docffy. Esta página tem finalidade **exclusivamente informativa** — ela não dispara nenhuma ação de negócio, não registra transações e não envia e-mails. Sua função é comunicar ao comprador que o pagamento foi submetido e que ele receberá o link de acesso por e-mail.

---

## 7. Regras de Split de Receita

### 7.1 Estrutura de taxas

Sobre cada venda realizada na plataforma, incidem dois conjuntos de taxas:

| Taxa | Modelo de cálculo |
|---|---|
| Taxa da plataforma Docffy | 2,91% do valor total + R$ 1,00 fixo |
| Taxa de processamento Stripe | 3,99% do valor total + R$ 0,39 fixo |

O valor líquido repassado ao criador é calculado da seguinte forma:

```
Valor ao criador = Valor total da venda
                 - Taxa Stripe (3,99% + R$ 0,39)
                 - Taxa Docffy (2,91% + R$ 1,00)
```

### 7.2 Condições para pagamento direto ao criador

O repasse automático ao criador ocorre apenas quando todas as condições abaixo forem atendidas simultaneamente:

| Condição | Descrição |
|---|---|
| Conta Stripe Connect vinculada | O criador possui uma conta Stripe Connect associada ao seu perfil |
| Onboarding concluído | O processo de cadastro na Stripe Connect foi finalizado pelo criador |
| Cobranças habilitadas | A Stripe confirmou que a conta do criador está apta a receber cobranças |

### 7.3 Pagamento retido na plataforma

Quando qualquer uma das condições da seção 7.2 não for satisfeita, o valor integral da venda fica retido na plataforma (`platform_only`). O repasse ao criador neste cenário depende de processo manual ou da conclusão do onboarding do criador junto à Stripe.

### 7.4 Tipo de split

O tipo de divisão de receita é registrado como `fixed` no momento da criação da transação pendente e atualizado para `percentage` na confirmação do pagamento via webhook, refletindo o modelo percentual de cálculo aplicado.

---

## 8. Regras de Confirmação de Compra

### 8.1 Princípio da fonte única de verdade

A confirmação de uma compra na Docffy ocorre **exclusivamente** mediante o recebimento do evento `checkout.session.completed` enviado pelo Stripe via webhook. Nenhuma outra ação do sistema — incluindo o redirecionamento para a página de sucesso — possui autoridade para confirmar, registrar ou alterar o estado de uma compra.

Este princípio garante que:

- Apenas compras efetivamente pagas sejam confirmadas;
- Não haja risco de ações duplicadas por recarga de página ou manipulação de URL;
- O sistema permaneça consistente mesmo em cenários de falha de rede ou timeout no lado do comprador.

### 8.2 Ações disparadas pelo webhook

Ao receber e validar o evento de confirmação do Stripe, o sistema executa, na seguinte ordem:

1. Localiza a sessão de compra associada ao evento;
2. Atualiza o status da compra de **pendente** para **confirmada**;
3. Atualiza o status da transação financeira de **pendente** para **concluída**;
4. Registra os valores definitivos do split de receita (percentual);
5. Envia o e-mail de entrega ao comprador com o link de acesso ao conteúdo.

### 8.3 Falha no pagamento

Caso o Stripe sinalize falha no pagamento, o status da transação é atualizado para **falha** e nenhuma das ações da seção 8.2 é executada. A compra permanece no estado pendente ou é marcada como cancelada, dependendo do evento recebido.

---

## 9. Regras de Entrega (Download)

### 9.1 Envio do link de acesso

O e-mail com o link de download é enviado **uma única vez**, automaticamente, imediatamente após a confirmação do pagamento via webhook. O envio nunca ocorre antes da confirmação nem na página de sucesso.

O e-mail contém:
- Confirmação da compra;
- Link de acesso ao e-book no formato `/purchase/download/{identificador-publico}`;
- Dados do criador para contato em caso de dúvidas.

### 9.2 URL de download

O link de acesso ao e-book utiliza o identificador público da compra (UUID v7), nunca o ID interno. Isso garante que as URLs sejam opacas, não sequenciais e não previsíveis.

### 9.3 Limite de downloads

| Configuração | Comportamento |
|---|---|
| Limite = -1 (padrão) | Downloads ilimitados — sem restrição de quantidade |
| Limite definido pelo criador | Número máximo de downloads permitidos para aquela compra |

Por padrão, toda compra permite downloads ilimitados. O criador pode configurar um limite diferente para seus produtos.

### 9.4 Expiração do acesso

Por padrão, o acesso ao e-book não expira. O campo de data de expiração é nulo na criação da compra, o que significa acesso permanente. Configurações de expiração são reservadas para regras específicas definidas pelo criador.

---

## 10. Regras de Gestão pelo Criador

### 10.1 Bloqueio e desbloqueio de acesso

O criador possui autonomia para bloquear ou desbloquear o acesso de um comprador ao e-book adquirido. Esta funcionalidade é conhecida como "congelamento" de acesso.

| Ação | Efeito |
|---|---|
| Bloquear (congelar) | O comprador não consegue mais realizar o download do e-book, mesmo que o pagamento tenha sido confirmado |
| Desbloquear | O acesso ao download é restabelecido para o comprador |

O bloqueio não cancela nem estorna a compra — ele apenas suspende temporariamente o acesso ao conteúdo.

### 10.2 Reenvio do link de download

O criador pode solicitar o reenvio do e-mail com o link de download para um comprador específico. Esta funcionalidade é destinada a situações em que o comprador não recebeu o e-mail original ou perdeu o acesso ao link.

O reenvio respeita as mesmas regras de acesso da seção 9 — se o acesso do comprador estiver bloqueado, o criador deve primeiro desbloquear antes de reenviar o link.

### 10.3 Histórico de vendas

O criador tem acesso ao histórico de todas as compras realizadas em seus e-books, incluindo dados do comprador (nome, e-mail, CPF parcialmente mascarado), data da compra, valor da transação e status atual do acesso. Este histórico é de uso exclusivo do criador e não é compartilhado entre criadores.

---

*Este documento reflete as regras de negócio implementadas e vigentes no sistema Docffy. Alterações nestas regras devem ser formalizadas por meio de uma nova User Story e aprovadas pelo Product Owner antes de qualquer implementação.*
