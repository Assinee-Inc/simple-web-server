document.addEventListener('DOMContentLoaded', function () {
  const form = document.getElementById('checkoutForm');
  const payButton = document.getElementById('payButton');
  const loadingSpinner = document.getElementById('loadingSpinner');

  function validateForm() {
    const name = document.getElementById('name').value || '';
    const cpf = document.getElementById('cpf').value || '';
    const birthdate = document.getElementById('birthdate').value || '';
    const email = document.getElementById('email').value || '';
    const phone = document.getElementById('phone').value || '';

    const isValid =
      name.length >= 3 &&
      cpf.replace(/\D/g, '').length === 11 &&
      birthdate.length === 10 &&
      email.includes('@') &&
      phone.length === 16;

    payButton.disabled = !isValid;
    return isValid;
  }

  validateForm();

  form.querySelectorAll('input').forEach(function (input) {
    input.addEventListener('input', validateForm);
  });

  form.addEventListener('submit', function (e) {
    e.preventDefault();
    if (!validateForm()) return;

    const formData = {
      name: document.getElementById('name').value.trim(),
      cpf: document.getElementById('cpf').value.replace(/\D/g, ''),
      birthdate: document.getElementById('birthdate').value,
      email: document.getElementById('email').value.trim(),
      phone: document.getElementById('phone').value.replace(/\D/g, ''),
      ebookId: document.getElementById('ebookId').value,
      csrfToken: document.getElementById('csrfToken').value,
    };

    loadingSpinner.style.display = 'flex';
    payButton.disabled = true;

    fetch('/api/validate-customer', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(formData),
    })
      .then(function (res) { return res.json(); })
      .then(function (response) {
        if (response.success) {
          createStripeSession(formData);
        } else {
          showError(response.error || 'Erro na validação dos dados');
        }
      })
      .catch(function () { showError('Erro na validação dos dados'); });
  });

  function createStripeSession(customerData) {
    fetch('/api/create-ebook-checkout', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(customerData),
    })
      .then(function (res) { return res.json(); })
      .then(function (response) {
        if (response.url) {
          window.location.href = response.url;
        } else {
          showError('Erro ao criar sessão de pagamento');
        }
      })
      .catch(function () { showError('Erro ao processar pagamento'); });
  }

  function showError(message) {
    loadingSpinner.style.display = 'none';
    payButton.disabled = false;
    alert('Erro: ' + message);
  }
});
