import React, { useEffect, useState } from "react";
import { Download, Pencil } from "lucide-react";
import { toast } from "sonner";
import { apiRequest, getAccessToken } from "@/lib/api";
import { formatPhone } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export function ClientListPage() {
  const [clients, setClients] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [term, setTerm] = useState("");

  const [editingClient, setEditingClient] = useState<any | null>(null);
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const loadClients = async (search = "") => {
    setLoading(true);
    try {
      const q = search ? `?term=${encodeURIComponent(search)}` : "";
      const res = await apiRequest<{ clients: any[] }>(`/client${q}`);
      setClients(res.clients || []);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao carregar clientes");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadClients();
  }, []);

  const handleUpdateClient = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingClient) return;
    const id =
      editingClient.public_id ||
      editingClient.PublicID ||
      String(editingClient.ID || "");
    setIsSaving(true);
    try {
      await apiRequest(`/client/${id}/update`, {
        method: "POST",
        body: JSON.stringify({
          email: email.trim(),
          phone: phone.replace(/\D/g, ""),
        }),
      });
      toast.success("Dados do cliente atualizados!");
      setEditingClient(null);
      loadClients(term);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao atualizar cliente");
    } finally {
      setIsSaving(false);
    }
  };

  const handleExportCSV = async () => {
    try {
      const token = getAccessToken();
      const res = await fetch("/client/export", {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (!res.ok) throw new Error("Falha ao exportar CSV");
      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "clientes_docffy.csv";
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao exportar CSV");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Clientes</h1>
          <p className="text-sm text-muted-foreground">
            Base de compradores vinculados aos seus e-books.
          </p>
        </div>
        <Button variant="outline" onClick={handleExportCSV}>
          <Download className="h-4 w-4" />
          Exportar CSV
        </Button>
      </div>

      <Card>
        <CardHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              loadClients(term);
            }}
            method="GET"
            action="/client"
            className="flex flex-col gap-3 sm:flex-row"
          >
            <Input
              name="term"
              placeholder="Buscar por nome, CPF ou e-mail..."
              value={term}
              onChange={(e) => setTerm(e.target.value)}
              autoCorrect="off"
              autoCapitalize="off"
              spellCheck={false}
              className="sm:max-w-sm"
            />
            <Button type="submit" variant="secondary">
              Buscar
            </Button>
          </form>
        </CardHeader>

        <CardContent>
          <Table ariaLabel="Tabela de clientes do criador">
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>E-mail</TableHead>
                <TableHead>CPF</TableHead>
                <TableHead>Telefone</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-8">
                    Carregando clientes...
                  </TableCell>
                </TableRow>
              ) : clients.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground py-8"
                  >
                    Nenhum cliente encontrado.
                  </TableCell>
                </TableRow>
              ) : (
                clients.map((c) => {
                  const id = c.public_id || c.PublicID || String(c.ID || "");
                  return (
                    <TableRow key={id}>
                      <TableCell className="font-medium">
                        {c.name || c.Name}
                      </TableCell>
                      <TableCell>{c.email || c.Email}</TableCell>
                      <TableCell className="font-mono text-xs">
                        {c.cpf || c.CPF}
                      </TableCell>
                      <TableCell>{c.phone || c.Phone || "—"}</TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => {
                            setEditingClient(c);
                            setEmail(c.email || c.Email || "");
                            setPhone(c.phone || c.Phone || "");
                          }}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
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
        open={Boolean(editingClient)}
        onOpenChange={(open) => !open && setEditingClient(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Atualizar Contato do Cliente</DialogTitle>
          </DialogHeader>
          <form
            onSubmit={handleUpdateClient}
            method="POST"
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="client-email">E-mail *</Label>
              <Input
                id="client-email"
                name="email"
                type="email"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="client-phone">Telefone *</Label>
              <Input
                id="client-phone"
                name="phone"
                type="text"
                inputMode="tel"
                value={phone}
                onChange={(e) => setPhone(formatPhone(e.target.value))}
                required
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditingClient(null)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={isSaving}>
                {isSaving ? "Salvando..." : "Salvar"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
