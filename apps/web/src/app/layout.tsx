import type { Metadata } from "next";
import Link from "next/link";
import type { ReactNode } from "react";

import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Antrean Review — Triage Klaim",
    template: "%s — Triage Klaim",
  },
  description: "Antrean alert indikatif untuk reviewer klaim P&C.",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="id">
      <body>
        <a className="skip-link" href="#main">
          Lompat ke konten utama
        </a>

        <header className="app-header">
          <div className="app-header__inner">
            <Link href="/queue" className="brand">
              <span className="brand__mark" aria-hidden="true">
                TK
              </span>
              <span className="brand__text">
                <span className="brand__name">Triage Klaim</span>
                <span className="brand__sub">Claim integrity review</span>
              </span>
            </Link>

            <nav aria-label="Navigasi utama" className="app-nav">
              <Link href="/queue" className="nav-link">
                Antrean Review
              </Link>
            </nav>

            <span className="env-badge">Thin slice · data sintetis</span>
          </div>
        </header>

        <main id="main" tabIndex={-1}>
          {children}
        </main>

        <footer className="app-footer">
          <p>
            Data sintetis untuk demo Phase 0/1. Alert bersifat indikatif dan
            tidak mengubah status klaim di sistem sumber.
          </p>
        </footer>
      </body>
    </html>
  );
}
