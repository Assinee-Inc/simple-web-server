document.addEventListener("DOMContentLoaded", function () {
  // CPF: 000.000.000-00
  document.querySelectorAll(".cpf").forEach(function (el) {
    IMask(el, { mask: "000.000.000-00" });
  });

  // Data: DD/MM/AAAA
  document.querySelectorAll(".date").forEach(function (el) {
    IMask(el, { mask: "00/00/0000" });
  });

  // Telefone com DDD: (00) 0 0000-0000
  document.querySelectorAll(".phone_with_ddd").forEach(function (el) {
    IMask(el, { mask: "(00) 0 0000-0000" });
  });

  // Dinheiro (money2): entrada da direita para esquerda (ex: digitar 1 → 0,01)
  document.querySelectorAll(".money2").forEach(function (el) {
    function formatMoney(digits) {
      var n = parseInt(digits || "0", 10);
      if (n === 0) return "";
      return (n / 100).toLocaleString("pt-BR", {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      });
    }

    // Valor inicial vindo do Go está em formato "29.90" — converter para "29,90"
    if (el.value.trim() !== "") {
      var cents = Math.round(parseFloat(el.value) * 100);
      el.value = cents > 0
        ? (cents / 100).toLocaleString("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 })
        : "";
    }

    el.addEventListener("input", function () {
      var digits = el.value.replace(/\D/g, "");
      el.value = formatMoney(digits);
    });
  });
});
