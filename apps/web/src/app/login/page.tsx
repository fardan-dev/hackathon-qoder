import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { getCurrentUser } from "@/lib/auth/session";

export const metadata: Metadata = {
  title: "Masuk",
};

type SearchParams = Promise<{ error?: string; next?: string }>;

const ERROR_COPY: Record<string, string> = {
  required: "Email dan password wajib diisi.",
  invalid: "Email atau password tidak sesuai.",
  rate: "Terlalu banyak percobaan masuk. Coba lagi dalam satu menit.",
  unavailable: "Layanan identitas tidak dapat dihubungi.",
};

function safeNext(value: string | undefined): string {
  if (!value || !value.startsWith("/") || value.startsWith("//")) {
    return "/queue";
  }
  return value;
}

export default async function LoginPage({
  searchParams,
}: {
  searchParams: SearchParams;
}) {
  const params = await searchParams;
  const next = safeNext(params.next);
  const user = await getCurrentUser();
  if (user) redirect(next);

  const error = params.error ? ERROR_COPY[params.error] : undefined;

  return (
    <div className="login-page">
      <section className="login-card" aria-labelledby="login-heading">
        <p className="eyebrow">Reviewer workspace</p>
        <h1 id="login-heading">Masuk ke Triage Klaim</h1>
        <p className="lede">
          Gunakan akun reviewer, supervisor, atau admin yang terdaftar. Sesi
          disimpan sebagai cookie HttpOnly; token tidak pernah disimpan di
          browser storage.
        </p>

        {error ? (
          <p className="login-error" role="alert">
            {error}
          </p>
        ) : null}

        <form className="login-form" action="/api/auth/login" method="post">
          <input type="hidden" name="next" value={next} />
          <div className="field">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="username"
              required
              autoFocus
            />
          </div>
          <div className="field">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
            />
          </div>
          <button className="button button--primary" type="submit">
            Masuk
          </button>
        </form>

        <p className="hint">
          Semua tampilan bersifat read-only. Keputusan dan perubahan status
          tetap dilakukan lewat API dengan otorisasi peran.
        </p>
      </section>
    </div>
  );
}
