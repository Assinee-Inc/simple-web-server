import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  BookOpen,
  ExternalLink,
  Eye,
  Pencil,
  Plus,
  Send,
  Trash2,
} from "lucide-react";
import { toast } from "sonner";
import { apiRequest } from "@/lib/api";
import { formatCPF, formatCurrencyBRL, formatPhone } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export interface EbookItem {
  ID?: number;
  PublicID?: string;
  public_id?: string;
  Title?: string;
  title?: string;
  Description?: string;
  description?: string;
  SalesPage?: string;
  sales_page?: string;
  AuthorName?: string;
  author_name?: string;
  Value?: number;
  value?: number;
  PromotionalValue?: number;
  promotional_value?: number;
  Status?: boolean;
  status?: boolean;
}

export function EbookListPage() {
  const [ebooks, setEbooks] = useState<EbookItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  // Create / Edit modal state
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [salesPage, setSalesPage] = useState("");
  const [authorName, setAuthorName] = useState("");
  const [value, setValue] = useState("49,90");
  const [promoValue, setPromoValue] = useState("");
  const [status, setStatus] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  // Send Purchase modal state
  const [sendModalEbookId, setSendModalEbookId] = useState<string | null>(null);
  const [buyerName, setBuyerName] = useState("");
  const [buyerCpf, setBuyerCpf] = useState("");
  const [buyerBirthdate, setBuyerBirthdate] = useState("1990-05-15");
  const [buyerEmail, setBuyerEmail] = useState("");
  const [buyerPhone, setBuyerPhone] = useState("");
  const [isSendingPurchase, setIsSendingPurchase] = useState(false);

  const loadEbooks = async (query = "") => {
    setLoading(true);
    try {
      const q = query ? `?title=${encodeURIComponent(query)}` : "";
      const res = await apiRequest<{ ebooks: EbookItem[] }>(`/ebook${q}`);
      setEbooks(res.ebooks || []);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao carregar e-books");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadEbooks();
  }, []);

  const openCreateModal = () => {
    setEditingId(null);
    setTitle("");
    setDescription("");
    setSalesPage("");
    setAuthorName("");
    setValue("49,90");
    setPromoValue("");
    setStatus(true);
    setModalOpen(true);
  };

  const openEditModal = (eb: EbookItem) => {
    const id = eb.public_id || eb.PublicID || String(eb.ID || "");
    setEditingId(id);
    setTitle(eb.title || eb.Title || "");
    setDescription(eb.description || eb.Description || "");
    setSalesPage(eb.sales_page || eb.SalesPage || "");
    setAuthorName(eb.author_name || eb.AuthorName || "");
    setValue(String(eb.value ?? eb.Value ?? "49,90").replace(".", ","));
    const pv = eb.promotional_value ?? eb.PromotionalValue;
    setPromoValue(pv ? String(pv).replace(".", ",") : "");
    setStatus(eb.status ?? eb.Status ?? true);
    setModalOpen(true);
  };

  const handleSaveEbook = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    try {
      const path = editingId ? `/ebook/update/${editingId}` : "/ebook/create";
      await apiRequest(path, {
        method: "POST",
        body: JSON.stringify({
          title,
          description,
          sales_page: salesPage,
          author_name: authorName,
          value,
          promotional_value: promoValue,
          status,
        }),
      });
      toast.success(
        editingId
          ? "E-book atualizado com sucesso!"
          : "E-book criado com sucesso!"
      );
      setModalOpen(false);
      loadEbooks(searchTerm);
    } catch (err: any) {
      toast.error(err?.message || "Falha ao salvar e-book");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Tem certeza que deseja excluir este e-book?")) return;
    try {
      await apiRequest(`/ebook/delete/${id}`, { method: "POST" });
      toast.success("E-book removido com sucesso!");
      loadEbooks(searchTerm);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao excluir e-book");
    }
  };

  const handleSendManualPurchase = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!sendModalEbookId) return;
    setIsSendingPurchase(true);
    try {
      await apiRequest(`/purchase/ebook/${sendModalEbookId}`, {
        method: "POST",
        body: JSON.stringify({
          clients: [
            {
              name: buyerName.trim(),
              cpf: buyerCpf.replace(/\D/g, ""),
              birthdate: buyerBirthdate,
              email: buyerEmail.trim(),
              phone: buyerPhone.replace(/\D/g, ""),
            },
          ],
        }),
      });
      toast.success("E-book enviado ao comprador com sucesso!");
      setSendModalEbookId(null);
      setBuyerName("");
      setBuyerCpf("");
      setBuyerEmail("");
      setBuyerPhone("");
    } catch (err: any) {
      toast.error(err?.message || "Erro ao registrar envio do e-book");
    } finally {
      setIsSendingPurchase(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Meus E-books</h1>
          <p className="text-sm text-muted-foreground">
            Gerencie seus produtos digitais, páginas de vendas e envios.
          </p>
        </div>
        <Button onClick={openCreateModal}>
          <Plus className="h-4 w-4" />
          Criar E-book
        </Button>
      </div>

      <Card>
        <CardHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              loadEbooks(searchTerm);
            }}
            method="GET"
            action="/ebook"
            className="flex flex-col gap-3 sm:flex-row"
          >
            <Input
              name="title"
              placeholder="Buscar e-book por título..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              autoCorrect="off"
              autoCapitalize="off"
              spellCheck={false}
              className="sm:max-w-sm"
            />
            <Button type="submit" variant="secondary">
              Filtrar
            </Button>
          </form>
        </CardHeader>

        <CardContent>
          <Table ariaLabel="Lista de e-books do criador">
            <TableHeader>
              <TableRow>
                <TableHead>Título</TableHead>
                <TableHead>Preço</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-8">
                    Carregando e-books...
                  </TableCell>
                </TableRow>
              ) : ebooks.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={4}
                    className="text-center text-muted-foreground py-8"
                  >
                    Nenhum e-book encontrado. Clique em "Criar E-book" para
                    começar.
                  </TableCell>
                </TableRow>
              ) : (
                ebooks.map((eb) => {
                  const id =
                    eb.public_id || eb.PublicID || String(eb.ID || "");
                  const isActive = eb.status ?? eb.Status ?? true;
                  const price = eb.value ?? eb.Value ?? 0;
                  const promo = eb.promotional_value ?? eb.PromotionalValue;

                  return (
                    <TableRow key={id}>
                      <TableCell>
                        <div className="font-medium">
                          {eb.title || eb.Title}
                        </div>
                        <div className="text-xs text-muted-foreground truncate max-w-md">
                          {eb.description || eb.Description}
                        </div>
                      </TableCell>
                      <TableCell>
                        {promo && promo > 0 ? (
                          <div>
                            <span className="font-semibold text-emerald-600">
                              {formatCurrencyBRL(promo)}
                            </span>{" "}
                            <span className="text-xs text-muted-foreground line-through">
                              {formatCurrencyBRL(price)}
                            </span>
                          </div>
                        ) : (
                          <span className="font-medium">
                            {formatCurrencyBRL(price)}
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={isActive ? "success" : "secondary"}>
                          {isActive ? "Ativo" : "Inativo"}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setSendModalEbookId(id)}
                            title="Enviar para comprador"
                          >
                            <Send className="h-3.5 w-3.5" />
                            Enviar
                          </Button>
                          <Button asChild variant="outline" size="sm">
                            <Link
                              to={`/sales/${id}`}
                              target="_blank"
                              rel="noreferrer"
                            >
                              <ExternalLink className="h-3.5 w-3.5" />
                              Página de Vendas
                            </Link>
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openEditModal(eb)}
                            title="Editar e-book"
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleDelete(id)}
                            title="Excluir e-book"
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
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

      {/* Modal Criar / Editar E-book */}
      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {editingId ? "Editar E-book" : "Criar Novo E-book"}
            </DialogTitle>
            <DialogDescription>
              Preencha as informações comerciais e de exibição da sua obra.
            </DialogDescription>
          </DialogHeader>

          <form
            onSubmit={handleSaveEbook}
            method="POST"
            action={editingId ? `/ebook/update/${editingId}` : "/ebook/create"}
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="ebook-title">Título do E-book *</Label>
              <Input
                id="ebook-title"
                name="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="ebook-description">Descrição curta *</Label>
              <Input
                id="ebook-description"
                name="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="ebook-sales-page">
                Texto da Página de Vendas *
              </Label>
              <textarea
                id="ebook-sales-page"
                name="sales_page"
                rows={4}
                className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={salesPage}
                onChange={(e) => setSalesPage(e.target.value)}
                required
              />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="space-y-1.5">
                <Label htmlFor="ebook-author">Nome do Autor</Label>
                <Input
                  id="ebook-author"
                  name="author_name"
                  value={authorName}
                  onChange={(e) => setAuthorName(e.target.value)}
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="ebook-value">Valor (R$) *</Label>
                <Input
                  id="ebook-value"
                  name="value"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="ebook-promo">Valor Promocional</Label>
                <Input
                  id="ebook-promo"
                  name="promotional_value"
                  value={promoValue}
                  onChange={(e) => setPromoValue(e.target.value)}
                  placeholder="Opcional"
                />
              </div>
            </div>

            <div className="flex items-center gap-2">
              <input
                id="ebook-status"
                name="status"
                type="checkbox"
                checked={status}
                onChange={(e) => setStatus(e.target.checked)}
                className="h-4 w-4 rounded border-input"
              />
              <Label htmlFor="ebook-status">
                Oferta ativa para vendas públicas
              </Label>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setModalOpen(false)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={isSaving}>
                {isSaving ? "Salvando..." : "Salvar E-book"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Modal Envio Manual de Compra */}
      <Dialog
        open={Boolean(sendModalEbookId)}
        onOpenChange={(open) => !open && setSendModalEbookId(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enviar E-book Protegido para Comprador</DialogTitle>
            <DialogDescription>
              Cria um vínculo de compra e envia o link de download com marca
              d'água dinâmica no CPF do comprador.
            </DialogDescription>
          </DialogHeader>

          <form
            onSubmit={handleSendManualPurchase}
            method="POST"
            action={`/purchase/ebook/${sendModalEbookId}`}
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="buyer-name">Nome completo do comprador *</Label>
              <Input
                id="buyer-name"
                name="name"
                value={buyerName}
                onChange={(e) => setBuyerName(e.target.value)}
                required
              />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="buyer-cpf">CPF do comprador *</Label>
                <Input
                  id="buyer-cpf"
                  name="cpf"
                  type="text"
                  inputMode="numeric"
                  autoCorrect="off"
                  autoCapitalize="off"
                  spellCheck={false}
                  value={buyerCpf}
                  onChange={(e) => setBuyerCpf(formatCPF(e.target.value))}
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="buyer-birthdate">Data de nascimento *</Label>
                <Input
                  id="buyer-birthdate"
                  name="birthdate"
                  type="date"
                  value={buyerBirthdate}
                  onChange={(e) => setBuyerBirthdate(e.target.value)}
                  required
                />
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="buyer-email">E-mail *</Label>
                <Input
                  id="buyer-email"
                  name="email"
                  type="email"
                  autoCorrect="off"
                  autoCapitalize="off"
                  spellCheck={false}
                  value={buyerEmail}
                  onChange={(e) => setBuyerEmail(e.target.value)}
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="buyer-phone">Telefone *</Label>
                <Input
                  id="buyer-phone"
                  name="phone"
                  type="text"
                  inputMode="tel"
                  value={buyerPhone}
                  onChange={(e) => setBuyerPhone(formatPhone(e.target.value))}
                  required
                />
              </div>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setSendModalEbookId(null)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={isSendingPurchase}>
                {isSendingPurchase ? "Enviando..." : "Enviar E-book"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
