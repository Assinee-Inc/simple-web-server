/**
 * Helper para upload direto de arquivos binários para o AWS S3 via URL pré-assinada (HTTP PUT).
 * O envio ocorre diretamente do navegador para a AWS, sem transitar pelo servidor backend.
 */
export async function uploadFileDirectToS3(
  uploadUrl: string,
  file: File,
  contentType?: string
): Promise<void> {
  const finalContentType = contentType || file.type || "application/pdf";

  const response = await fetch(uploadUrl, {
    method: "PUT",
    headers: {
      "Content-Type": finalContentType,
    },
    body: file,
  });

  if (!response.ok) {
    throw new Error(
      `Falha no upload direto para a AWS S3 (Status ${response.status}: ${
        response.statusText || "Erro no envio"
      }). Verifique as permissões de CORS ou se a URL expirou.`
    );
  }
}
