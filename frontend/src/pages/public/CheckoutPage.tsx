import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { AlertCircle, Lock, ShieldCheck } from "lucide-react";
import { apiRequest } from "@/lib/api";
import {
  formatCPF,
  formatCurrencyBRL,
  formatDateBR,
  formatPhone,
} from "@/lib/utils";
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
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export function CheckoutPage() {
  const { id } = useParams<{ id: string }>();
  const [ebook, setEbook] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [name, setName] = useState("");
  const [cpf, setCpf] = useState("");
  const [birthDate, setBirthDate] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!id) return;
    apiRequest(`/checkout/${id}`)
      .then((res) => setEbook(res.ebook || res))
      .catch((err) =>
        setErrorMessage(err?.message || "Oferta não encontrada para checkout")
      )
      .finally(() => setLoading(false));
  }, [id]);

  const handleCheckout = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;
    setErrorMessage(null);
    setIsSubmitting(true);

    try {
      const cleanCpf = cpf.replace(/\D/g, "");
      await apiRequest("/api/validate-customer", {
        method: "POST",
        body: JSON.stringify({
          ebook_id: id,
          cpf: cleanCpf,
          birthdate: birthDate.trim(),
          email: email.trim(),
        }),
      });

      const checkoutRes = await apiRequest<{
        checkout_url?: string;
        url?: string;
      }>("/api/create-ebook-checkout", {
        method: "POST",
        body: JSON.stringify({
          ebook_id: id,
          name: name.trim(),
          cpf: cleanCpf,
          birthdate: birthDate.trim(),
          email: email.trim(),
          phone: phone.replace(/\D/g, ""),
        }),
      });

      const redirectUrl = checkoutRes.checkout_url || checkoutRes.url;
      if (redirectUrl) {
        window.location.href = redirectUrl;
      }
    } catch (err: any) {
      setErrorMessage(
        err?.message || "Não foi possível validar seus dados para o pagamento."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center p-6">
        <p className="text-muted-foreground">Carregando checkout seguro...</p>
      </div>
    );
  }

  const price = ebook?.value ?? ebook?.Value ?? 0;
  const promo = ebook?.promotional_value ?? ebook?.PromotionalValue;
  const finalPrice = promo && promo > 0 ? promo : price;

  return (
    <div className="min-h-screen bg-muted/30 py-10 px-4">
      <div className="mx-auto max-w-lg space-y-6">
        <div className="flex items-center justify-center gap-2">
          <ShieldCheck className="h-6 w-6 text-primary" />
          <span className="text-lg font-bold">Checkout Seguro Docffy</span>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>{ebook?.title || ebook?.Title || "E-book"}</CardTitle>
            <CardDescription>
              Total a pagar:{" "}
              <strong className="text-foreground">
                {formatCurrencyBRL(finalPrice)}
              </strong>
            </CardDescription>
          </CardHeader>

          <CardContent>
            {errorMessage && (
              <Alert variant="destructive" className="mb-4">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Atenção</AlertTitle>
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
            )}

            <form
              onSubmit={handleCheckout}
              method="POST"
              action="/api/create-ebook-checkout"
              className="space-y-4"
            >
              <div className="space-y-1.5">
                <Label htmlFor="checkout-name">Nome completo *</Label>
                <Input
                  id="checkout-name"
                  name="name"
                  autoComplete="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                />
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <Label htmlFor="checkout-cpf">CPF (para marca d'água) *</Label>
                  <Input
                    id="checkout-cpf"
                    name="cpf"
                    type="text"
                    inputMode="numeric"
                    autoCorrect="off"
                    autoCapitalize="off"
                    spellCheck={false}
                    value={cpf}
                    onChange={(e) => setCpf(formatCPF(e.target.value))}
                    placeholder="000.000.000-00"
                    required
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="checkout-bday">
                    Data de nascimento (18+) *
                  </Label>
                  <Input
                    id="checkout-bday"
                    name="birthdate"
                    type="text"
                    inputMode="numeric"
                    autoComplete="bday"
                    value={birthDate}
                    onChange={(e) => setBirthDate(formatDateBR(e.target.value))}
                    placeholder="DD/MM/AAAA"
                    required
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="checkout-email">
                  E-mail (onde você receberá o e-book) *
                </Label>
                <Input
                  id="checkout-email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  autoCorrect="off"
                  autoCapitalize="off"
                  spellCheck={false}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="checkout-phone">Telefone / WhatsApp *</Label>
                <Input
                  id="checkout-phone"
                  name="phone"
                  type="text"
                  inputMode="tel"
                  autoComplete="tel"
                  value={phone}
                  onChange={(e) => setPhone(formatPhone(e.target.value))}
                  required
                />
              </div>

              <Button
                type="submit"
                size="lg"
                className="w-full"
                disabled={isSubmitting}
              >
                <Lock className="h-4 w-4" />
                {isSubmitting
                  ? "Validando e redirecionando..."
                  : `Pagar ${formatCurrencyBRL(finalPrice)}`}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
