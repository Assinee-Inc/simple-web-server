import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { BookOpen, ShieldCheck, ShoppingCart } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { formatCurrencyBRL } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export function SalesPage({ preview = false }: { preview?: boolean }) {
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    const endpoint = preview ? `/ebook/preview/${id}` : `/sales/${id}`;
    apiRequest(endpoint)
      .then((res) => setData(res))
      .catch((err) => setError(err?.message || "Oferta não encontrada"))
      .finally(() => setLoading(false));
  }, [id, preview]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center p-6">
        <p className="text-muted-foreground">Carregando oferta...</p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center p-6 text-center">
        <h1 className="text-xl font-bold">Oferta indisponível</h1>
        <p className="mt-2 text-sm text-muted-foreground">{error}</p>
      </div>
    );
  }

  const ebook = data.ebook || data;
  const creator = data.creator || {};
  const title = ebook.title || ebook.Title || "E-book";
  const description = ebook.description || ebook.Description || "";
  const salesContent = ebook.sales_page || ebook.SalesPage || "";
  const price = ebook.value ?? ebook.Value ?? 0;
  const promo = ebook.promotional_value ?? ebook.PromotionalValue;
  const finalPrice = promo && promo > 0 ? promo : price;

  return (
    <div className="min-h-screen bg-muted/30 py-12 px-4">
      <div className="mx-auto max-w-3xl space-y-6">
        {preview && (
          <div className="rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-2 text-center text-xs font-medium text-amber-700">
            Modo Pré-visualização do Criador
          </div>
        )}

        <Card className="overflow-hidden">
          <CardHeader className="border-b bg-card pb-6">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <BookOpen className="h-4 w-4" />
              <span>
                Por{" "}
                {ebook.author_name ||
                  ebook.AuthorName ||
                  creator.social_name ||
                  creator.name ||
                  "Autor Docffy"}
              </span>
            </div>
            <CardTitle className="mt-2 text-3xl font-bold">{title}</CardTitle>
            <CardDescription className="text-base">
              {description}
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-6 pt-6">
            <div className="prose prose-sm max-w-none whitespace-pre-line text-foreground/90 leading-relaxed">
              {salesContent}
            </div>

            <div className="flex flex-col items-start justify-between gap-4 rounded-xl border bg-muted/40 p-6 sm:flex-row sm:items-center">
              <div>
                <span className="text-xs uppercase text-muted-foreground">
                  Investimento único
                </span>
                <div className="flex items-baseline gap-2">
                  <span className="text-3xl font-bold text-foreground">
                    {formatCurrencyBRL(finalPrice)}
                  </span>
                  {promo && promo > 0 && (
                    <span className="text-sm text-muted-foreground line-through">
                      {formatCurrencyBRL(price)}
                    </span>
                  )}
                </div>
              </div>

              <Button asChild size="lg">
                <Link to={`/checkout/${id}`}>
                  <ShoppingCart className="h-4 w-4" />
                  Comprar Agora ({formatCurrencyBRL(finalPrice)})
                </Link>
              </Button>
            </div>
          </CardContent>

          <CardFooter className="border-t bg-muted/20 py-4 text-xs text-muted-foreground flex items-center justify-between">
            <div className="flex items-center gap-1.5">
              <ShieldCheck className="h-4 w-4 text-emerald-600" />
              <span>
                Entrega digital imediata protegida pelo Docffy (PDF com marca
                d'água)
              </span>
            </div>
            <Badge variant="outline">Compra 100% Segura</Badge>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
