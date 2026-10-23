import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { AlertCircle, Download, FileText, ShieldCheck } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function DownloadPage() {
  const { hashId } = useParams<{ hashId: string }>();
  const [data, setData] = useState<any | null>(null);
  const [error, setError] = useState<{ code?: string; message: string } | null>(
    null
  );
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!hashId) return;
    apiRequest(`/purchase/download/${hashId}`)
      .then((res) => setData(res))
      .catch((err: any) =>
        setError({
          code: err?.code,
          message:
            err?.message ||
            "Não foi possível acessar os arquivos desta compra.",
        })
      )
      .finally(() => setLoading(false));
  }, [hashId]);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-muted/30 p-4">
      <div className="mb-6 flex items-center gap-2">
        <ShieldCheck className="h-6 w-6 text-primary" />
        <span className="font-bold">Docffy — Entrega Digital Protegida</span>
      </div>

      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>
            {data?.ebook_title ||
              data?.ebook?.title ||
              data?.ebook?.Title ||
              "Download do seu E-book"}
          </CardTitle>
          <CardDescription>
            Cada arquivo baixado recebe marca d'água dinâmica com sua
            identificação de comprador.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {loading ? (
            <p className="text-sm text-muted-foreground">
              Verificando credencial de download...
            </p>
          ) : error ? (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>
                {error.code || "Download Indisponível"}
              </AlertTitle>
              <AlertDescription>{error.message}</AlertDescription>
            </Alert>
          ) : (
            <>
              <div className="flex items-center justify-between rounded-lg border bg-muted/40 p-3 text-xs">
                <span>Downloads utilizados:</span>
                <Badge variant="secondary">
                  {data?.downloads_used ?? 0} /{" "}
                  {data?.download_limit === -1
                    ? "Ilimitado"
                    : data?.download_limit ?? 5}
                </Badge>
              </div>

              <div className="space-y-2">
                {(data?.files || data?.ebook?.Files || []).map((file: any) => {
                  const fileId =
                    file.public_id || file.PublicID || String(file.ID || "");
                  const fileName =
                    file.name ||
                    file.Name ||
                    file.original_name ||
                    "Arquivo PDF";
                  return (
                    <div
                      key={fileId}
                      className="flex items-center justify-between rounded-lg border p-3"
                    >
                      <div className="flex items-center gap-2.5">
                        <FileText className="h-5 w-5 text-primary" />
                        <span className="text-sm font-medium">{fileName}</span>
                      </div>
                      <Button asChild size="sm">
                        <a
                          href={`/purchase/download/${hashId}?file_id=${encodeURIComponent(
                            fileId
                          )}`}
                        >
                          <Download className="h-4 w-4" />
                          Baixar PDF
                        </a>
                      </Button>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
