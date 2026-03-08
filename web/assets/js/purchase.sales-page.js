function buyNow() {
  const previewBanner = document.querySelector('.preview-banner');
  if (previewBanner) {
    alert('Esta é uma visualização da página de vendas. A funcionalidade de compra não está disponível no modo preview.');
    return;
  }
  const ebookId = document.querySelector('[data-ebook-id]').dataset.ebookId;
  window.location.href = '/checkout/' + ebookId;
}
