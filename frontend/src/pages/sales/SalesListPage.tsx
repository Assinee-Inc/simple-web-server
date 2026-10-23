import React, { useEffect, useState } from "react";
import { Lock, Mail, Receipt, Unlock } from "lucide-react";
import { toast } from "sonner";
import { apiRequest } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export function SalesListPage() {
  const [purchases, setPurchases] = useState<any[]>([]);
  const [transactionMap, setTransactionMap] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState(true);
  const [clientEmail, setClientEmail] = useState("");

  // Transaction detail modal
  const [selectedTx, setSelectedTx] = useState<any | null>(null);

  const loadSales = async (emailFilter = "") => {
    setLoading(true);
    try {
      const q = emailFilter
        ? `?client_email=${encodeURIComponent(emailFilter)}`
        : "";
      const res = await apiRequest<{
        purchases: any[];
        transaction_map?: Record<string, any>;
      }>(`/purchase/sales${q}`);
      setPurchases(res.purchases || []);
      setTransactionMap(res.transaction_map || {});
    } catch (err: any) {
      toast.error(err?.message || "Erro ao carregar vendas");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSales();
  }, []);

  const toggleBlockDownload = async (purchaseId: string, blocked: boolean) => {
    const endpoint = blocked
      ? "/purchase/sales/unblock-download"
      : "/purchase/sales/block-download";
    try {
      await apiRequest(endpoint, {
        method: "POST",
        body: JSON.stringify({ purchase_id: String(purchaseId) }),
      });
      toast.success(
        blocked ? "Download desbloqueado!" : "Download bloqueado!"
      );
      loadSales(clientEmail);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao alterar bloqueio de download");
    }
  };

  const handleResendLink = async (purchaseId: string) => {
    try {
      await apiRequest("/purchase/sales/resend-link", {
        method: "POST",
        body: JSON.stringify({ purchase_id: String(purchaseId) }),
      });
      toast.success("Link de download reenviado ao comprador!");
    } catch (err: any) {
      toast.error(err?.message || "Falha ao reenviar link");
    }
  };

  const openTransactionDetail = async (txId: string | number) => {
    try {
      const res = await apiRequest(`/transactions/detail?id=${txId}`);
      setSelectedTx(res.transaction || res);
    } catch (err: any) {
      toast.error(err?.message || "Não foi possível carregar a transação");
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">
          Vendas & Entregas Realizadas
        </h1>
        <p className="text-sm text-muted-foreground">
          Acompanhe compras, gerencie permissões de download e reenvie links de
          acesso.
        </p>
      </div>

      <Card>
        <CardHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              loadSales(clientEmail);
            }}
            method="GET"
            action="/purchase/sales"
            className="flex flex-col gap-3 sm:flex-row"
          >
            <Input
              name="client_email"
              placeholder="Filtrar por e-mail do comprador..."
              value={clientEmail}
              onChange={(e) => setClientEmail(e.target.value)}
              autoCorrect="off"
              autoCapitalize="off"
              spellCheck={false}
              className="sm:max-w-sm"
            />
            <Button type="submit" variant="secondary">
              Filtrar Vendas
            </Button>
          </form>
        </CardHeader>

        <CardContent>
          <Table ariaLabel="Histórico de vendas do criador">
            <TableHeader>
              <TableRow>
                <TableHead>Comprador</TableHead>
                <TableHead>E-book</TableHead>
                <TableHead>Downloads</TableHead>
                <TableHead>Status Acesso</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-8">
                    Carregando vendas...
                  </TableCell>
                </TableRow>
              ) : purchases.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground py-8"
                  >
                    Nenhuma venda encontrada.
                  </TableCell>
                </TableRow>
              ) : (
                purchases.map((p) => {
                  const id = String(p.ID || p.id || p.public_id || "");
                  const blocked = Boolean(
                    p.BlockDownload ?? p.block_download ?? false
                  );
                  const clientName =
                    p.Client?.Name || p.client?.name || "Comprador";
                  const email = p.Client?.Email || p.client?.email || "";
                  const ebookTitle =
                    p.Ebook?.Title || p.ebook?.title || "E-book";
                  const used = p.DownloadsUsed ?? p.downloads_used ?? 0;
                  const limit = p.DownloadLimit ?? p.download_limit ?? 5;
                  const tx = transactionMap[id];

                  return (
                    <TableRow key={id}>
                      <TableCell>
                        <div className="font-medium">{clientName}</div>
                        <div className="text-xs text-muted-foreground">
                          {email}
                        </div>
                      </TableCell>
                      <TableCell className="font-medium">
                        {ebookTitle}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {used} / {limit === -1 ? "∞" : limit}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={blocked ? "destructive" : "success"}>
                          {blocked ? "Bloqueado" : "Liberado"}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleResendLink(id)}
                          >
                            <Mail className="h-3.5 w-3.5" />
                            Reenviar Link
                          </Button>
                          <Button
                            variant={blocked ? "secondary" : "outline"}
                            size="sm"
                            onClick={() => toggleBlockDownload(id, blocked)}
                          >
                            {blocked ? (
                              <>
                                <Unlock className="h-3.5 w-3.5" />
                                Desbloquear
                              </>
                            ) : (
                              <>
                                <Lock className="h-3.5 w-3.5" />
                                Bloquear
                              </>
                            )}
                          </Button>
                          {tx && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                openTransactionDetail(
                                  tx.ID || tx.id || tx.public_id
                                )
                              }
                            >
                              <Receipt className="h-3.5 w-3.5" />
                              Transação
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog
        open={Boolean(selectedTx)}
        onOpenChange={(open) => !open && setSelectedTx(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Detalhes da Transação</DialogTitle>
            <DialogDescription>
              Informações financeiras e status do gateway de pagamento.
            </DialogDescription>
          </DialogHeader>
          {selectedTx && (
            <div className="space-y-2 text-sm">
              <div className="flex justify-between border-b py-1.5">
                <span className="text-muted-foreground">Status:</span>
                <Badge variant="secondary">
                  {selectedTx.status || selectedTx.Status || "completed"}
                </Badge>
              </div>
              <div className="flex justify-between border-b py-1.5">
                <span className="text-muted-foreground">Stripe ID:</span>
                <span className="font-mono text-xs">
                  {selectedTx.stripe_payment_intent_id ||
                    selectedTx.StripePaymentIntentID ||
                    "—"}
                </span>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
