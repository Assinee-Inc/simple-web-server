document.addEventListener('DOMContentLoaded', function () {
    const previewModal = document.getElementById('filePreviewModal');
    const previewContainer = document.getElementById('previewContainer');

    previewModal.addEventListener('close', function () {
        previewContainer.innerHTML = '<p class="text-base-content/60 mt-5">Carregando preview...</p>';
    });
});

function openPreviewModal(fileURL, fileType) {
    const previewModal = document.getElementById('filePreviewModal');
    const previewContainer = document.getElementById('previewContainer');

    // Limpa o container de previews anteriores
    previewContainer.innerHTML = '';

    // Define o título do modal
    const modalTitle = document.getElementById('filePreviewModalLabel');
    modalTitle.textContent = 'Visualizando: ' + fileType;

    let previewContent;

    // Lógica para selecionar o tipo de preview com base no Content-Type
    if (fileType.startsWith('image/')) {
        // Se for imagem, usa a tag <img>
        previewContent = `<img src="${fileURL}" class="rounded max-h-full max-w-full" alt="Pré-visualização da Imagem">`;

    } else if (fileType === 'application/pdf') {
        // Se for PDF, usa a tag <object> para um melhor suporte nativo do navegador
        previewContent = `
        <object
            data="${fileURL}"
            type="application/pdf"
            width="100%"
            height="100%"
            style="border: none;">
            <p>O seu navegador não suporta a visualização de PDFs.
            <a href="${fileURL}" target="_blank">Clique aqui para baixar.</a></p>
        </object>`;
    } else if (fileType.startsWith('video/')) {
        // Se for vídeo, usa a tag <video>
        previewContent = `
        <video controls width="100%" height="auto" style="max-height: 100%;">
            <source src="${fileURL}" type="${fileType}">
            Seu navegador não suporta o elemento de vídeo.
        </video>`;
    } else {
        // Fallback para outros tipos (como arquivos de texto ou desconhecidos)
        previewContent = `
        <p class="text-error mt-5">
            Não é possível exibir o 'preview' nativamente para o tipo de arquivo:
            <strong>${fileType}</strong>.
        </p>
        <a href="${fileURL}" target="_blank" class="btn btn-success mt-3">Tentar Abrir em Nova Aba</a>`;
    }

    // Injeta o conteúdo no modal
    previewContainer.innerHTML = previewContent;

    // Abre o modal
    previewModal.showModal();
}
