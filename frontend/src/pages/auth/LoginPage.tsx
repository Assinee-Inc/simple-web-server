import React, { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ShieldCheck, KeyRound, AlertCircle } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/contexts/AuthContext";
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
import { Separator } from "@/components/ui/separator";

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: { credential: string }) => void;
          }) => void;
          renderButton: (
            parent: HTMLElement,
            options: Record<string, unknown>
          ) => void;
        };
      };
    };
  }
}

export function LoginPage() {
  const { loginWithGoogle, user, onboardingRequired } = useAuth();
  const navigate = useNavigate();
  const googleButtonRef = useRef<HTMLDivElement>(null);

  const [manualIdToken, setManualIdToken] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || "";

  useEffect(() => {
    if (user) {
      navigate(onboardingRequired ? "/onboarding" : "/dashboard", {
        replace: true,
      });
    }
  }, [user, onboardingRequired, navigate]);

  const handleGoogleCredential = async (idToken: string) => {
    setErrorMessage(null);
    setIsSubmitting(true);
    try {
      const result = await loginWithGoogle(idToken);
      toast.success("Autenticado com sucesso via Google!");
      if (result.onboarding_required) {
        navigate("/onboarding");
      } else {
        navigate("/dashboard");
      }
    } catch (err: any) {
      setErrorMessage(
        err?.message || "Falha ao autenticar com a credencial Google."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  useEffect(() => {
    if (!googleClientId || !googleButtonRef.current) return;

    const interval = setInterval(() => {
      if (window.google?.accounts?.id && googleButtonRef.current) {
        clearInterval(interval);
        window.google.accounts.id.initialize({
          client_id: googleClientId,
          callback: (response) => {
            if (response.credential) {
              handleGoogleCredential(response.credential);
            }
          },
        });
        window.google.accounts.id.renderButton(googleButtonRef.current, {
          theme: "outline",
          size: "large",
          width: 360,
          text: "continue_with",
          shape: "rectangular",
        });
      }
    }, 200);

    return () => clearInterval(interval);
  }, [googleClientId]);

  const handleManualSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!manualIdToken.trim()) {
      setErrorMessage("Informe o id_token emitido pelo Google.");
      return;
    }
    await handleGoogleCredential(manualIdToken.trim());
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-muted/30 p-4">
      <div className="mb-8 flex items-center gap-3">
        <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
          <ShieldCheck className="h-6 w-6" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Docffy</h1>
          <p className="text-xs text-muted-foreground">
            Entrega protegida de e-books com marca d'água dinâmica
          </p>
        </div>
      </div>

      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Acesse sua conta</CardTitle>
          <CardDescription>
            Autenticação 100% sem senhas utilizando sua conta Google verificada.
            Contas existentes com o mesmo e-mail são vinculadas automaticamente.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {errorMessage && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Erro de autenticação</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}

          {googleClientId ? (
            <div className="flex justify-center py-2">
              <div ref={googleButtonRef} />
            </div>
          ) : (
            <div className="rounded-lg border border-dashed bg-muted/40 p-3 text-xs text-muted-foreground">
              Para exibir o botão automático do Google Identity Services, defina{" "}
              <code className="font-mono text-foreground">
                VITE_GOOGLE_CLIENT_ID
              </code>{" "}
              no arquivo <code className="font-mono">.env</code> do frontend.
            </div>
          )}

          <div className="relative flex items-center justify-center">
            <Separator />
            <span className="bg-card px-2 text-xs text-muted-foreground uppercase">
              ou use um ID Token (Dev / Postman)
            </span>
          </div>

          <form
            onSubmit={handleManualSubmit}
            method="POST"
            action="/api/auth/google"
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="google-id-token">Google OIDC ID Token</Label>
              <Input
                id="google-id-token"
                name="id_token"
                type="text"
                placeholder="eyJhbGciOiJSUzI1NiIs..."
                value={manualIdToken}
                onChange={(e) => {
                  setErrorMessage(null);
                  setManualIdToken(e.target.value);
                }}
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                required
              />
              <p className="text-xs text-muted-foreground">
                Cole o <code className="font-mono">id_token</code> gerado no
                Google OAuth Playground ou Postman.
              </p>
            </div>

            <Button type="submit" className="w-full" disabled={isSubmitting}>
              <KeyRound className="h-4 w-4" />
              {isSubmitting
                ? "Verificando credencial Google..."
                : "Continuar com Google ID Token"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
