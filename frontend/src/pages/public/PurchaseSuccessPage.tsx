import React, { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { CheckCircle2, Mail } from "lucide-react";
import { apiRequest } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function PurchaseSuccessPage() {
  const [searchParams] = useSearchParams();
  const sessionId = searchParams.get("session_id") || "";
  const [data, setData] = useState<any | null>(null);

  useEffect(() => {
    if (!sessionId) return;
    apiRequest(`/purchase/success?session_id=${encodeURIComponent(sessionId)}`)
      .then((res) => setData(res))
      .catch(() => {});
  }, [sessionId]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-md text-center">
        <CardHeader>
          <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600">
            <CheckCircle2 className="h-7 w-7" />
          </div>
          <CardTitle>Compra Confirmada!</CardTitle>
          <CardDescription>
            Seu pagamento foi processado com sucesso.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-sm text-muted-foreground">
          <div className="flex items-center justify-center gap-2 rounded-lg border bg-muted/40 p-3 text-foreground">
            <Mail className="h-4 w-4 text-primary" />
            <span>
              Enviamos o link exclusivo de download para{" "}
              <strong>{data?.customer_email || "o seu e-mail"}</strong>.
            </span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
