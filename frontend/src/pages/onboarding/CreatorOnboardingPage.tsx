import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { AlertCircle, CheckCircle2, ShieldCheck } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/contexts/AuthContext";
import { apiRequest, ApiError } from "@/lib/api";
import { formatCPF, formatDateBR, formatPhone } from "@/lib/utils";
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

export function CreatorOnboardingPage() {
  const { user, refreshSession, logout } = useAuth();
  const navigate = useNavigate();

  const [name, setName] = useState(user?.username || "");
  const [socialName, setSocialName] = useState("");
  const [cpf, setCpf] = useState("");
  const [birthDate, setBirthDate] = useState("");
  const [phone, setPhone] = useState("");
  const [termsAccepted, setTermsAccepted] = useState(false);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setFieldErrors({});
    setIsSubmitting(true);

    try {
      await apiRequest("/api/account/onboarding", {
        method: "POST",
        body: JSON.stringify({
          name: name.trim(),
          social_name: socialName.trim(),
          cpf: cpf.replace(/\D/g, ""),
          birthdate: birthDate.trim(),
          phone: phone.replace(/\D/g, ""),
          terms_accepted: termsAccepted,
        }),
      });

      await refreshSession();
      toast.success("Perfil de criador concluído com sucesso!");
      navigate("/dashboard");
    } catch (err: any) {
      if (err instanceof ApiError) {
        setErrorMessage(err.message);
        if (err.details) {
          setFieldErrors(err.details);
        }
      } else {
        setErrorMessage("Não foi possível concluir o cadastro de criador.");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-muted/30 p-4">
      <div className="mb-6 flex w-full max-w-lg items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <span className="font-bold">Docffy — Etapa 2 de 2</span>
        </div>
        <Button variant="ghost" size="sm" onClick={() => logout()}>
          Trocar conta ({user?.email})
        </Button>
      </div>

      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>Complete seu perfil de Criador</CardTitle>
          <CardDescription>
            Sua conta Google (<strong>{user?.email}</strong>) já está
            autenticada. Informe seus dados fiscais para liberar a criação de
            e-books e recebimentos.
          </CardDescription>
        </CardHeader>

        <CardContent>
          {errorMessage && (
            <Alert variant="destructive" className="mb-6">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Verifique os dados informados</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}

          <form
            onSubmit={handleSubmit}
            method="POST"
            action="/api/account/onboarding"
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="creator-name">Nome completo (civil) *</Label>
              <Input
                id="creator-name"
                name="name"
                type="text"
                autoComplete="name"
                value={name}
                onChange={(e) => {
                  setFieldErrors((prev) => ({ ...prev, name: "" }));
                  setName(e.target.value);
                }}
                placeholder="Ex.: Maria da Silva"
                required
              />
              {fieldErrors.name && (
                <p className="text-xs text-destructive">{fieldErrors.name}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="creator-social-name">
                Nome social / Nome artístico (opcional)
              </Label>
              <Input
                id="creator-social-name"
                name="social_name"
                type="text"
                value={socialName}
                onChange={(e) => setSocialName(e.target.value)}
                placeholder="Como seu nome aparecerá para compradores"
              />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="creator-cpf">CPF *</Label>
                <Input
                  id="creator-cpf"
                  name="cpf"
                  type="text"
                  inputMode="numeric"
                  autoCorrect="off"
                  autoCapitalize="off"
                  spellCheck={false}
                  value={cpf}
                  onChange={(e) => {
                    setFieldErrors((prev) => ({ ...prev, cpf: "" }));
                    setCpf(formatCPF(e.target.value));
                  }}
                  placeholder="000.000.000-00"
                  required
                />
                {fieldErrors.cpf && (
                  <p className="text-xs text-destructive">{fieldErrors.cpf}</p>
                )}
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="creator-birthdate">
                  Data de nascimento (18+) *
                </Label>
                <Input
                  id="creator-birthdate"
                  name="birthdate"
                  type="text"
                  inputMode="numeric"
                  autoComplete="bday"
                  value={birthDate}
                  onChange={(e) => {
                    setFieldErrors((prev) => ({ ...prev, birthdate: "" }));
                    setBirthDate(formatDateBR(e.target.value));
                  }}
                  placeholder="DD/MM/AAAA"
                  required
                />
                {fieldErrors.birthdate && (
                  <p className="text-xs text-destructive">
                    {fieldErrors.birthdate}
                  </p>
                )}
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="creator-phone">Telefone / WhatsApp *</Label>
              <Input
                id="creator-phone"
                name="phone"
                type="text"
                inputMode="tel"
                autoComplete="tel"
                value={phone}
                onChange={(e) => {
                  setFieldErrors((prev) => ({ ...prev, phone: "" }));
                  setPhone(formatPhone(e.target.value));
                }}
                placeholder="(11) 99999-9999"
                required
              />
              {fieldErrors.phone && (
                <p className="text-xs text-destructive">{fieldErrors.phone}</p>
              )}
            </div>

            <div className="flex items-start gap-3 rounded-lg border p-3">
              <input
                id="creator-terms"
                name="terms_accepted"
                type="checkbox"
                className="mt-1 h-4 w-4 rounded border-input"
                checked={termsAccepted}
                onChange={(e) => setTermsAccepted(e.target.checked)}
                required
              />
              <Label
                htmlFor="creator-terms"
                className="text-xs leading-relaxed text-muted-foreground cursor-pointer"
              >
                Declaro que sou maior de 18 anos e concordo com os Termos de Uso
                e Política de Privacidade do Docffy.
              </Label>
            </div>

            <Button type="submit" className="w-full" disabled={isSubmitting}>
              <CheckCircle2 className="h-4 w-4" />
              {isSubmitting
                ? "Validando CPF e criando perfil..."
                : "Concluir Cadastro de Criador"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
