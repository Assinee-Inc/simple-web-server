import React, { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { CheckCircle2, CreditCard, ExternalLink, RefreshCw } from "lucide-react";
import { toast } from "sonner";
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

interface StripeStatusResponse {
  has_stripe_account?: boolean;
  onboarding_completed?: boolean;
  payouts_enabled?: boolean;
  charges_enabled?: boolean;
  needs_onboarding?: boolean;
  remediation_link?: string;
  onboarding_url?: string;
}

export function StripeConnectPage() {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState<StripeStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [startingOnboarding, setStartingOnboarding] = useState(false);

  const loadStatus = async () => {
    setLoading(true);
    try {
      const res = await apiRequest<StripeStatusResponse>(
        "/stripe-connect/status"
      );
      setStatus(res);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao consultar status Stripe Connect");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const statusParam = searchParams.get("status");
    if (statusParam === "complete") {
      apiRequest("/stripe-connect/complete")
        .then(() => {
          toast.success("Onboarding do Stripe Connect sincronizado com sucesso!");
          window.history.replaceState({}, "", window.location.pathname);
          loadStatus();
        })
        .catch(() => {
          loadStatus();
        });
    } else {
      loadStatus();
    }
  }, [searchParams]);

  const handleStartOnboarding = async () => {
    setStartingOnboarding(true);
    try {
      const returnUrl = `${window.location.origin}/stripe-connect?status=complete`;
      const refreshUrl = `${window.location.origin}/stripe-connect?status=refresh`;
      const res = await apiRequest<{ onboarding_url?: string }>(
        `/stripe-connect/onboard?return_url=${encodeURIComponent(returnUrl)}&refresh_url=${encodeURIComponent(refreshUrl)}`
      );
      if (res.onboarding_url) {
        window.location.href = res.onboarding_url;
      } else {
        toast.info("Link de onboarding gerado.");
        loadStatus();
      }
    } catch (err: any) {
      toast.error(err?.message || "Erro ao iniciar onboarding Stripe Connect");
    } finally {
      setStartingOnboarding(false);
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            Recebimentos Stripe Connect
          </h1>
          <p className="text-sm text-muted-foreground">
            Configure sua conta bancária conectada para receber repasses
            automáticos das vendas.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={loadStatus}>
          <RefreshCw className="h-4 w-4" />
          Atualizar
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <CreditCard className="h-5 w-5" />
            Status da Conta de Recebimento
          </CardTitle>
          <CardDescription>
            Integração direta com a Stripe para processamento de pagamentos e
            transferências.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {loading ? (
            <p className="text-sm text-muted-foreground">
              Consultando conta Stripe...
            </p>
          ) : (
            <>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground">
                    Onboarding
                  </div>
                  <div className="mt-1">
                    <Badge
                      variant={
                        status?.onboarding_completed ? "success" : "warning"
                      }
                    >
                      {status?.onboarding_completed ? "Concluído" : "Pendente"}
                    </Badge>
                  </div>
                </div>

                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground">Cobranças</div>
                  <div className="mt-1">
                    <Badge
                      variant={status?.charges_enabled ? "success" : "secondary"}
                    >
                      {status?.charges_enabled ? "Habilitadas" : "Inativas"}
                    </Badge>
                  </div>
                </div>

                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground">Saques</div>
                  <div className="mt-1">
                    <Badge
                      variant={status?.payouts_enabled ? "success" : "secondary"}
                    >
                      {status?.payouts_enabled ? "Habilitados" : "Inativos"}
                    </Badge>
                  </div>
                </div>
              </div>

              {status?.onboarding_completed &&
              status?.charges_enabled &&
              status?.payouts_enabled ? (
                <div className="flex items-center gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4 text-sm text-emerald-700 dark:text-emerald-300">
                  <CheckCircle2 className="h-5 w-5 shrink-0" />
                  <span>
                    Sua conta Stripe Connect está ativa e pronta para receber
                    pagamentos!
                  </span>
                </div>
              ) : (
                <div className="pt-2">
                  <Button
                    onClick={handleStartOnboarding}
                    disabled={startingOnboarding}
                  >
                    <ExternalLink className="h-4 w-4" />
                    {startingOnboarding
                      ? "Redirecionando para a Stripe..."
                      : "Concluir Cadastro na Stripe Connect"}
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
