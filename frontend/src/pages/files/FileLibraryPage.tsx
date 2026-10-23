import React, { useEffect, useState } from "react";
import { FileText, Pencil, Trash2, Upload } from "lucide-react";
import { toast } from "sonner";
import { apiRequest } from "@/lib/api";
import { uploadFileDirectToS3 } from "@/lib/s3-upload";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
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

interface FileItem {
  ID?: number;
  PublicID?: string;
  public_id?: string;
  Name?: string;
  name?: string;
  Description?: string;
  description?: string;
  OriginalName?: string;
  original_name?: string;
  FileType?: string;
  file_type?: string;
  FileSize?: number;
  file_size?: number;
}

export function FileLibraryPage() {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");

  // Upload dialog
  const [uploadOpen, setUploadOpen] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [uploadDescription, setUploadDescription] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadStage, setUploadStage] = useState<
    "idle" | "requesting_url" | "uploading_s3" | "completing"
  >("idle");
  const [uploadError, setUploadError] = useState<string | null>(null);

  // Edit dialog
  const [editingFile, setEditingFile] = useState<FileItem | null>(null);
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  const loadFiles = async (q = "") => {
    setLoading(true);
    try {
      const query = q ? `?search=${encodeURIComponent(q)}` : "";
      const res = await apiRequest<{ files: FileItem[] }>(`/file${query}`);
      setFiles(res.files || []);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao carregar biblioteca de arquivos");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFiles();
  }, []);

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile) {
      toast.error("Selecione um arquivo PDF para enviar.");
      return;
    }
    setUploadError(null);
    setIsUploading(true);
    setUploadStage("requesting_url");
    try {
      // 1. Obter URL pré-assinada do S3 junto ao backend
      const presignRes = await apiRequest<{
        success: boolean;
        upload_url: string;
        key: string;
      }>("/file/upload-url", {
        method: "POST",
        body: JSON.stringify({
          file_name: selectedFile.name,
          file_size: selectedFile.size,
          content_type: selectedFile.type || "application/pdf",
        }),
      });

      if (!presignRes.upload_url || !presignRes.key) {
        throw new Error("URL pré-assinada não retornada pelo servidor.");
      }

      // 2. Upload direto do binário para a AWS S3 (sem passar pelo backend)
      setUploadStage("uploading_s3");
      await uploadFileDirectToS3(
        presignRes.upload_url,
        selectedFile,
        selectedFile.type || "application/pdf"
      );

      // 3. Confirmar e registrar metadados no backend
      setUploadStage("completing");
      await apiRequest("/file/complete-upload", {
        method: "POST",
        body: JSON.stringify({
          key: presignRes.key,
          name: uploadName || selectedFile.name,
          description: uploadDescription,
          original_name: selectedFile.name,
          size: selectedFile.size,
          content_type: selectedFile.type || "application/pdf",
        }),
      });

      toast.success("Arquivo enviado diretamente para a AWS S3 com sucesso!");
      setUploadOpen(false);
      setUploadName("");
      setUploadDescription("");
      setSelectedFile(null);
      setUploadError(null);
      loadFiles(search);
    } catch (err: any) {
      setUploadError(err?.message || "Falha no upload do arquivo.");
      toast.error(err?.message || "Falha no upload do arquivo");
    } finally {
      setIsUploading(false);
      setUploadStage("idle");
    }
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingFile) return;
    const id =
      editingFile.public_id ||
      editingFile.PublicID ||
      String(editingFile.ID || "");
    setIsUpdating(true);
    try {
      await apiRequest(`/file/${id}/update`, {
        method: "POST",
        body: JSON.stringify({
          name: editName,
          description: editDescription,
        }),
      });
      toast.success("Metadados do arquivo atualizados!");
      setEditingFile(null);
      loadFiles(search);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao atualizar arquivo");
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm("Deseja remover este arquivo permanentemente?")) return;
    try {
      await apiRequest(`/file/${id}/delete`, { method: "POST" });
      toast.success("Arquivo excluído com sucesso!");
      loadFiles(search);
    } catch (err: any) {
      toast.error(err?.message || "Erro ao excluir arquivo");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            Biblioteca de Arquivos (S3)
          </h1>
          <p className="text-sm text-muted-foreground">
            Envie e gerencie os PDFs originais que receberão marca d'água na
            entrega.
          </p>
        </div>
        <Button onClick={() => setUploadOpen(true)}>
          <Upload className="h-4 w-4" />
          Enviar Arquivo
        </Button>
      </div>

      <Card>
        <CardHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              loadFiles(search);
            }}
            method="GET"
            action="/file"
            className="flex flex-col gap-3 sm:flex-row"
          >
            <Input
              name="search"
              placeholder="Buscar arquivo por nome..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
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
          <Table ariaLabel="Biblioteca de arquivos do criador">
            <TableHeader>
              <TableRow>
                <TableHead>Arquivo</TableHead>
                <TableHead>Descrição</TableHead>
                <TableHead>Tipo</TableHead>
                <TableHead className="text-right">Ações</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-8">
                    Carregando arquivos...
                  </TableCell>
                </TableRow>
              ) : files.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={4}
                    className="text-center text-muted-foreground py-8"
                  >
                    Nenhum arquivo enviado ainda.
                  </TableCell>
                </TableRow>
              ) : (
                files.map((f) => {
                  const id = f.public_id || f.PublicID || String(f.ID || "");
                  return (
                    <TableRow key={id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground" />
                          <span>{f.name || f.Name}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {f.description || f.Description || "—"}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {f.file_type || f.FileType || "PDF"}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => {
                              setEditingFile(f);
                              setEditName(f.name || f.Name || "");
                              setEditDescription(
                                f.description || f.Description || ""
                              );
                            }}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleDelete(id)}
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

      {/* Upload Dialog */}
      <Dialog
        open={uploadOpen}
        onOpenChange={(open) => {
          setUploadOpen(open);
          if (!open) setUploadError(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enviar Novo Arquivo</DialogTitle>
            <DialogDescription>
              Selecione o arquivo PDF que será armazenado de forma segura no AWS
              S3.
            </DialogDescription>
          </DialogHeader>

          {uploadError && (
            <div className="p-3 text-sm rounded-md bg-destructive/10 text-destructive border border-destructive/20">
              {uploadError}
            </div>
          )}

          <form
            onSubmit={handleUpload}
            method="POST"
            action="/file/upload"
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="upload-name">Nome de exibição *</Label>
              <Input
                id="upload-name"
                name="name"
                value={uploadName}
                onChange={(e) => setUploadName(e.target.value)}
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="upload-description">Descrição</Label>
              <Input
                id="upload-description"
                name="description"
                value={uploadDescription}
                onChange={(e) => setUploadDescription(e.target.value)}
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="upload-file">Arquivo PDF *</Label>
              <Input
                id="upload-file"
                name="file"
                type="file"
                accept=".pdf,application/pdf"
                onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                required
              />
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setUploadOpen(false)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={isUploading}>
                {isUploading
                  ? uploadStage === "requesting_url"
                    ? "Gerando link seguro..."
                    : uploadStage === "uploading_s3"
                    ? "Enviando para a AWS..."
                    : "Finalizando cadastro..."
                  : uploadError
                  ? "Tentar Novamente"
                  : "Enviar Arquivo"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog
        open={Boolean(editingFile)}
        onOpenChange={(open) => !open && setEditingFile(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Editar Arquivo</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleUpdate} method="POST" className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="edit-file-name">Nome *</Label>
              <Input
                id="edit-file-name"
                name="name"
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="edit-file-desc">Descrição</Label>
              <Input
                id="edit-file-desc"
                name="description"
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditingFile(null)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={isUpdating}>
                {isUpdating ? "Salvando..." : "Salvar"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
