import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  ExternalLink,
  FileText,
  Paperclip,
  Pencil,
  Plus,
  Send,
  Trash2,
} from "lucide-react";
import { toast } from "sonner";
import { apiRequest } from "@/lib/api";
import { cn, formatCPF, formatCurrencyBRL, formatPhone } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
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

export interface FileItem {
  id?: number;
  ID?: number;
  public_id?: string;
  PublicID?: string;
  name?: string;
  Name?: string;
  original_name?: string;
  OriginalName?: string;
  file_type?: string;
  FileType?: string;
  file_size?: number;
  FileSize?: number;
}

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
  Files?: FileItem[];
  files?: FileItem[];
}

function formatFileSize(bytes?: number): string {
  if (!bytes || bytes <= 0) return "";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

export function EbookListPage() {
  const [ebooks, setEbooks] = useState<EbookItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  // Available files from the creator library
  const [availableFiles, setAvailableFiles] = useState<FileItem[]>([]);
  const [loadingFiles, setLoadingFiles] = useState(false);

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

  // File management inside modal
  const [selectedFileIds, setSelectedFileIds] = useState<string[]>([]);
  const [attachedFiles, setAttachedFiles] = useState<FileItem[]>([]);
  const [removingFileId, setRemovingFileId] = useState<string | null>(null);

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

  const loadAvailableFiles = async () => {
    setLoadingFiles(true);
    try {
      const res = await apiRequest<{ files: FileItem[] }>("/file?page_size=100");
      setAvailableFiles(res.files || []);
    } catch (err: any) {
      console.error("Erro ao carregar arquivos da biblioteca:", err);
    } finally {
      setLoadingFiles(false);
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
    setSelectedFileIds([]);
    setAttachedFiles([]);
    setModalOpen(true);
    loadAvailableFiles();
  };

  const openEditModal = async (eb: EbookItem) => {
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
    setSelectedFileIds([]);

    const initialFiles = eb.files || eb.Files || [];
    setAttachedFiles(initialFiles);
    setModalOpen(true);
    loadAvailableFiles();

    try {
      const res = await apiRequest<{ ebook?: EbookItem; available_files?: FileItem[] }>(
        `/ebook/edit/${id}`
      );
      if (res.ebook) {
        setAttachedFiles(res.ebook.files || res.ebook.Files || []);
      }
      if (res.available_files) {
        setAvailableFiles(res.available_files);
      }
    } catch {
      // Usa dados existentes se requisição falhar
    }
  };

  const toggleSelectFile = (fileId: string) => {
    setSelectedFileIds((prev) =>
      prev.includes(fileId) ? prev.filter((id) => id !== fileId) : [...prev, fileId]
    );
  };

  const handleRemoveFileFromEbook = async (fileId: string) => {
    if (!editingId) return;
    if (attachedFiles.length <= 1) {
      toast.error("O e-book deve ter pelo menos um arquivo vinculado.");
      return;
    }
    if (!window.confirm("Deseja realmente desvincular este arquivo deste e-book?")) {
      return;
    }
    setRemovingFileId(fileId);
    try {
      await apiRequest(`/ebook/${editingId}/remove-file/${fileId}`, {
        method: "POST",
      });
      toast.success("Arquivo desvinculado com sucesso!");
      setAttachedFiles((prev) =>
        prev.filter((f) => (f.public_id || f.PublicID || String(f.id || f.ID || "")) !== fileId)
      );
      loadEbooks(searchTerm);
      loadAvailableFiles();
    } catch (err: any) {
      toast.error(err?.message || "Erro ao desvincular arquivo");
    } finally {
      setRemovingFileId(null);
    }
  };

  const handleSaveEbook = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingId && selectedFileIds.length === 0) {
      toast.error("Por favor, selecione pelo menos um arquivo para vincular ao e-book.");
      return;
    }
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
          new_files: selectedFileIds,
          files: selectedFileIds,
          file_ids: selectedFileIds,
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

  // Computa arquivos da biblioteca que ainda não estão vinculados ao e-book em edição
  const attachedIdSet = new Set(
    attachedFiles.map((f) => f.public_id || f.PublicID || String(f.id || f.ID || ""))
  );
  const unattachedAvailableFiles = availableFiles.filter(
    (f) => !attachedIdSet.has(f.public_id || f.PublicID || String(f.id || f.ID || ""))
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Meus E-books</h1>
          <p className="text-sm text-muted-foreground">
            Gerencie seus produtos digitais, arquivos vinculados e páginas de vendas.
          </p>
        </div>
        <Button onClick={openCreateModal}>
          <Plus className="h-4 w-4 mr-1.5" />
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
                <TableHead>Arquivos Vinculados</TableHead>
                <TableHead>Preço</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-8">
                    Carregando e-books...
                  </TableCell>
                </TableRow>
              ) : ebooks.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
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
                  const ebFiles = eb.files || eb.Files || [];

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
                        {ebFiles.length > 0 ? (
                          <div className="flex flex-col gap-1 max-w-[220px]">
                            <div className="flex items-center gap-1.5 text-xs font-medium text-foreground">
                              <Paperclip className="h-3.5 w-3.5 text-primary shrink-0" />
                              <span>
                                {ebFiles.length}{" "}
                                {ebFiles.length === 1
                                  ? "arquivo"
                                  : "arquivos"}
                              </span>
                            </div>
                            <div className="text-[11px] text-muted-foreground truncate">
                              {ebFiles
                                .map((f) => f.name || f.Name || f.original_name)
                                .join(", ")}
                            </div>
                          </div>
                        ) : (
                          <Badge
                            variant="destructive"
                            className="text-[11px] font-normal"
                          >
                            Sem arquivo vinculado
                          </Badge>
                        )}
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
                            <Send className="h-3.5 w-3.5 mr-1" />
                            Enviar
                          </Button>
                          <Button asChild variant="outline" size="sm">
                            <Link
                              to={`/sales/${id}`}
                              target="_blank"
                              rel="noreferrer"
                            >
                              <ExternalLink className="h-3.5 w-3.5 mr-1" />
                              Vendas
                            </Link>
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openEditModal(eb)}
                            title="Editar e-book e arquivos"
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
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editingId ? "Editar E-book" : "Criar Novo E-book"}
            </DialogTitle>
            <DialogDescription>
              Configure os dados comerciais e vincule os arquivos da sua biblioteca.
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
                rows={3}
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

            {/* SEÇÃO: VÍNCULO DE ARQUIVOS */}
            <div className="space-y-3 rounded-lg border p-4 bg-muted/20">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4 text-primary" />
                  <Label className="text-sm font-semibold">
                    {editingId
                      ? "Arquivos do Conteúdo Digital"
                      : "Vincular Arquivos da Biblioteca *"}
                  </Label>
                </div>
                <Button
                  asChild
                  variant="ghost"
                  size="sm"
                  className="text-xs text-muted-foreground hover:text-foreground h-7"
                >
                  <Link to="/library/files" target="_blank">
                    <ExternalLink className="h-3 w-3 mr-1" />
                    Abrir Biblioteca de Arquivos
                  </Link>
                </Button>
              </div>

              {/* Se estiver editando: exibir arquivos já vinculados com opção de desvincular */}
              {editingId && (
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground">
                    Arquivos atualmente associados a este e-book:
                  </p>
                  {attachedFiles.length === 0 ? (
                    <div className="p-3 text-center text-xs text-amber-700 bg-amber-50 rounded-md border border-amber-200">
                      Nenhum arquivo associado. Selecione arquivos abaixo para vincular.
                    </div>
                  ) : (
                    <div className="space-y-1.5">
                      {attachedFiles.map((f) => {
                        const fId =
                          f.public_id ||
                          f.PublicID ||
                          String(f.id || f.ID || "");
                        const fName =
                          f.name ||
                          f.Name ||
                          f.original_name ||
                          f.OriginalName ||
                          "Arquivo sem nome";
                        const fType = f.file_type || f.FileType || "PDF";
                        const fSize = f.file_size || f.FileSize;
                        const isRemoving = removingFileId === fId;

                        return (
                          <div
                            key={fId}
                            className="flex items-center justify-between p-2 rounded-md border bg-background text-sm"
                          >
                            <div className="flex items-center gap-2 min-w-0">
                              <FileText className="h-4 w-4 text-primary shrink-0" />
                              <span className="font-medium truncate max-w-xs sm:max-w-sm">
                                {fName}
                              </span>
                              <Badge
                                variant="outline"
                                className="text-[10px] shrink-0"
                              >
                                {fType}
                              </Badge>
                              {fSize ? (
                                <span className="text-xs text-muted-foreground shrink-0">
                                  {formatFileSize(fSize)}
                                </span>
                              ) : null}
                            </div>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              className="text-destructive hover:text-destructive hover:bg-destructive/10 h-7 text-xs shrink-0"
                              disabled={isRemoving}
                              onClick={() => handleRemoveFileFromEbook(fId)}
                            >
                              <Trash2 className="h-3.5 w-3.5 mr-1" />
                              {isRemoving ? "Desvinculando..." : "Desvincular"}
                            </Button>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              )}

              {/* Seleção de arquivos da biblioteca */}
              <div className="space-y-2">
                <p className="text-xs font-medium text-muted-foreground">
                  {editingId
                    ? "Vincular arquivos adicionais da biblioteca:"
                    : "Selecione os arquivos que farão parte deste e-book:"}
                </p>

                {loadingFiles ? (
                  <div className="py-4 text-center text-xs text-muted-foreground">
                    Carregando arquivos da biblioteca...
                  </div>
                ) : unattachedAvailableFiles.length === 0 ? (
                  <div className="p-3 text-center text-xs text-muted-foreground bg-background rounded-md border">
                    {availableFiles.length === 0 ? (
                      <div>
                        Nenhum arquivo encontrado na biblioteca.{" "}
                        <Link
                          to="/library/files"
                          className="text-primary underline font-medium"
                        >
                          Fazer upload na biblioteca
                        </Link>
                      </div>
                    ) : (
                      "Todos os arquivos da sua biblioteca já estão vinculados a este e-book."
                    )}
                  </div>
                ) : (
                  <div className="max-h-48 overflow-y-auto space-y-1.5 border rounded-md p-2 bg-background">
                    {unattachedAvailableFiles.map((f) => {
                      const fId =
                        f.public_id ||
                        f.PublicID ||
                        String(f.id || f.ID || "");
                      const fName =
                        f.name ||
                        f.Name ||
                        f.original_name ||
                        f.OriginalName ||
                        "Arquivo sem nome";
                      const fType = f.file_type || f.FileType || "PDF";
                      const fSize = f.file_size || f.FileSize;
                      const isChecked = selectedFileIds.includes(fId);

                      return (
                        <label
                          key={fId}
                          htmlFor={`select-file-${fId}`}
                          className={cn(
                            "flex items-center justify-between p-2 rounded-md border cursor-pointer text-sm transition-colors",
                            isChecked
                              ? "border-primary bg-primary/5"
                              : "border-border hover:bg-muted/40"
                          )}
                        >
                          <div className="flex items-center gap-2.5 min-w-0">
                            <input
                              id={`select-file-${fId}`}
                              type="checkbox"
                              checked={isChecked}
                              onChange={() => toggleSelectFile(fId)}
                              className="h-4 w-4 rounded border-input text-primary"
                            />
                            <span className="font-medium truncate max-w-xs sm:max-w-md">
                              {fName}
                            </span>
                            <Badge
                              variant="outline"
                              className="text-[10px] shrink-0"
                            >
                              {fType}
                            </Badge>
                            {fSize ? (
                              <span className="text-xs text-muted-foreground shrink-0">
                                {formatFileSize(fSize)}
                              </span>
                            ) : null}
                          </div>
                        </label>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2 pt-1">
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
