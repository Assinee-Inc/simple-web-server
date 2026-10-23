import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  BookOpen,
  Download,
  Plus,
  Send,
  TrendingUp,
  Users,
} from "lucide-react";
import { apiRequest } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

interface DashboardData {
  total_ebooks: number;
  total_sent_ebooks: number;
  total_clients: number;
  top_ebooks?: Array<{
    Title?: string;
    title?: string;
    SalesCount?: number;
    sales_count?: number;
  }>;
  top_clients?: Array<{
    Name?: string;
    name?: string;
    Email?: string;
    email?: string;
    PurchaseCount?: number;
    purchase_count?: number;
  }>;
  top_downloaded_ebooks?: Array<{
    Title?: string;
    title?: string;
    DownloadCount?: number;
    download_count?: number;
  }>;
}

export function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiRequest<DashboardData>("/dashboard")
      .then((res) => setData(res))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-sm text-muted-foreground">
            Visão geral das suas publicações, envios protegidos e clientes.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button asChild variant="outline">
            <Link to="/files">Gerenciar Arquivos</Link>
          </Button>
          <Button asChild>
            <Link to="/ebooks">
              <Plus className="h-4 w-4" />
              Novo E-book
            </Link>
          </Button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total de E-books
            </CardTitle>
            <BookOpen className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {loading ? "..." : data?.total_ebooks ?? 0}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Catálogo ativo na sua biblioteca
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              E-books Enviados / Vendidos
            </CardTitle>
            <Send className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {loading ? "..." : data?.total_sent_ebooks ?? 0}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Entregas protegidas com marca d'água
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total de Clientes
            </CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {loading ? "..." : data?.total_clients ?? 0}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Compradores cadastrados
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Rankings */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <TrendingUp className="h-4 w-4" />
              E-books Mais Vendidos
            </CardTitle>
            <CardDescription>
              Publicações com maior volume de vendas/envios
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table ariaLabel="E-books mais vendidos">
              <TableHeader>
                <TableRow>
                  <TableHead>Título</TableHead>
                  <TableHead className="text-right">Vendas</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.top_ebooks && data.top_ebooks.length > 0 ? (
                  data.top_ebooks.map((eb, idx) => (
                    <TableRow key={idx}>
                      <TableCell className="font-medium">
                        {eb.Title || eb.title || "Sem título"}
                      </TableCell>
                      <TableCell className="text-right">
                        <Badge variant="secondary">
                          {eb.SalesCount ?? eb.sales_count ?? 0}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={2}
                      className="text-center text-muted-foreground py-6"
                    >
                      Nenhuma venda registrada ainda.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Download className="h-4 w-4" />
              E-books Mais Baixados
            </CardTitle>
            <CardDescription>
              Downloads protegidos com marca d'água realizados
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table ariaLabel="E-books mais baixados">
              <TableHeader>
                <TableRow>
                  <TableHead>Título</TableHead>
                  <TableHead className="text-right">Downloads</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.top_downloaded_ebooks &&
                data.top_downloaded_ebooks.length > 0 ? (
                  data.top_downloaded_ebooks.map((eb, idx) => (
                    <TableRow key={idx}>
                      <TableCell className="font-medium">
                        {eb.Title || eb.title || "Sem título"}
                      </TableCell>
                      <TableCell className="text-right">
                        <Badge variant="secondary">
                          {eb.DownloadCount ?? eb.download_count ?? 0}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell
                      colSpan={2}
                      className="text-center text-muted-foreground py-6"
                    >
                      Nenhum download registrado ainda.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
