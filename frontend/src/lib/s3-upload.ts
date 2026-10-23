export class S3UploadError extends Error {
  isExpired: boolean;
  status: number;

  constructor(message: string, status: number, isExpired: boolean = false) {
    super(message);
    this.name = "S3UploadError";
    this.status = status;
    this.isExpired = isExpired;
  }
}

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

  let response: Response;
  try {
    response = await fetch(uploadUrl, {
      method: "PUT",
      headers: {
        "Content-Type": finalContentType,
      },
      body: file,
    });
  } catch (err: any) {
    throw new S3UploadError(
      `Falha de conexão com a AWS S3 (${err?.message || "Erro de rede"}). Verifique sua conexão e tente novamente.`,
      0,
      false
    );
  }

  if (!response.ok) {
    const errorBody = await response.text().catch(() => "");
    const isExpired =
      response.status === 403 &&
      (errorBody.includes("Request has expired") ||
        errorBody.includes("ExpiredToken") ||
        errorBody.includes("AccessDenied"));

    if (isExpired) {
      throw new S3UploadError(
        "A autorização de upload expirou (limite de 60 segundos). Clique em 'Tentar Novamente' para gerar um novo link seguro.",
        response.status,
        true
      );
    }

    throw new S3UploadError(
      `Falha no upload direto para a AWS S3 (Status ${response.status}: ${
        response.statusText || "Erro no envio"
      }). Verifique as permissões de CORS ou se a URL expirou.`,
      response.status,
      false
    );
  }
}
