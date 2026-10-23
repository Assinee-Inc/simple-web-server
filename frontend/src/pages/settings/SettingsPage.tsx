import React, { useEffect, useState } from "react";
import { Save } from "lucide-react";
import { toast } from "sonner";
import { apiRequest } from "@/lib/api";
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

export function SettingsPage() {
  const { refreshSession } = useAuth();
  const [socialName, setSocialName] = useState("");
  const [facebookPixelId, setFacebookPixelId] = useState("");
  const [savingSocial, setSavingSocial] = useState(false);
  const [savingPixel, setSavingPixel] = useState(false);

  useEffect(() => {
    apiRequest<{
      creator?: { social_name?: string; facebook_pixel_id?: string };
      social_name?: string;
      facebook_pixel_id?: string;
    }>("/settings")
      .then((res) => {
        setSocialName(
          res.social_name || res.creator?.social_name || ""
        );
        setFacebookPixelId(
          res.facebook_pixel_id || res.creator?.facebook_pixel_id || ""
        );
      })
      .catch(() => {});
  }, []);

  const handleSaveSocialName = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingSocial(true);
    try {
      await apiRequest("/settings/social-name", {
        method: "POST",
        body: JSON.stringify({ social_name: socialName.trim() }),
      });
      await refreshSession();
      toast.success("Nome social atualizado com sucesso!");
    } catch (err: any) {
      toast.error(err?.message || "Erro ao atualizar nome social");
    } finally {
      setSavingSocial(false);
    }
  };

  const handleSavePixel = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingPixel(true);
    try {
      await apiRequest("/settings/facebook-pixel", {
        method: "POST",
        body: JSON.stringify({ facebook_pixel_id: facebookPixelId.trim() }),
      });
      toast.success("Meta Pixel ID atualizado com sucesso!");
    } catch (err: any) {
      toast.error(err?.message || "Erro ao atualizar Meta Pixel ID");
    } finally {
      setSavingPixel(false);
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Configurações</h1>
        <p className="text-sm text-muted-foreground">
          Personalize sua identidade pública de autor e rastreamento de
          conversões.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Nome Social / Exibição</CardTitle>
          <CardDescription>
            Nome apresentado aos compradores nas páginas de vendas e e-mails.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={handleSaveSocialName}
            method="POST"
            action="/settings/social-name"
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="settings-social-name">Nome de exibição</Label>
              <Input
                id="settings-social-name"
                name="social_name"
                value={socialName}
                onChange={(e) => setSocialName(e.target.value)}
                placeholder="Ex.: Prof. Maria Silva"
              />
            </div>
            <Button type="submit" disabled={savingSocial}>
              <Save className="h-4 w-4" />
              {savingSocial ? "Salvando..." : "Salvar Nome Social"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            Rastreamento Meta Pixel (Facebook Ads)
          </CardTitle>
          <CardDescription>
            Informe o ID numérico de 10 a 20 dígitos do seu Meta Pixel para
            rastrear PageView e Purchase nas páginas de vendas.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={handleSavePixel}
            method="POST"
            action="/settings/facebook-pixel"
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="settings-pixel-id">Meta Pixel ID</Label>
              <Input
                id="settings-pixel-id"
                name="facebook_pixel_id"
                type="text"
                inputMode="numeric"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                value={facebookPixelId}
                onChange={(e) =>
                  setFacebookPixelId(e.target.value.replace(/\D/g, ""))
                }
                placeholder="Ex.: 123456789012345"
              />
            </div>
            <Button type="submit" disabled={savingPixel}>
              <Save className="h-4 w-4" />
              {savingPixel ? "Salvando..." : "Salvar Meta Pixel"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
