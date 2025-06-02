import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/hooks/useAuth";
import { Toaster } from "react-hot-toast";

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "DanteGPU - Decentralized GPU Rental Platform",
  description: "Professional GPU rental platform powered by Solana blockchain and dGPU tokens. Rent high-performance GPUs for AI, machine learning, rendering, and cryptocurrency mining.",
  keywords: "GPU rental, AI computing, machine learning, cryptocurrency mining, Solana blockchain, dGPU tokens",
  authors: [{ name: "DanteGPU Team" }],
  creator: "DanteGPU",
  publisher: "DanteGPU",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  icons: {
    icon: "/dantegpu-logo.png",
    shortcut: "/dantegpu-logo.png",
    apple: "/dantegpu-logo.png",
  },
  manifest: "/manifest.json",
  openGraph: {
    title: "DanteGPU - Decentralized GPU Rental Platform",
    description: "Professional GPU rental platform powered by Solana blockchain and dGPU tokens",
    url: "https://dantegpu.com",
    siteName: "DanteGPU",
    images: [
      {
        url: "/dantegpu-logo.png",
        width: 1200,
        height: 630,
        alt: "DanteGPU Logo",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "DanteGPU - Decentralized GPU Rental Platform",
    description: "Professional GPU rental platform powered by Solana blockchain and dGPU tokens",
    images: ["/dantegpu-logo.png"],
    creator: "@DanteGPU",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  verification: {
    google: "google-verification-code",
    yandex: "yandex-verification-code",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${inter.variable}`}>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link 
          href="https://fonts.googleapis.com/css2?family=Hubot+Sans:ital,wght@0,200..900;1,200..900&display=swap" 
          rel="stylesheet" 
        />
      </head>
      <body className={`font-hubot-sans antialiased`}>
        <AuthProvider>
          <div className="min-h-full">
            {children}
          </div>
          <Toaster 
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: "#363636",
                color: "#fff",
              },
              success: {
                iconTheme: {
                  primary: "#4ade80",
                  secondary: "#fff",
                },
              },
              error: {
                iconTheme: {
                  primary: "#ef4444",
                  secondary: "#fff",
                },
              },
            }}
          />
        </AuthProvider>
      </body>
    </html>
  );
}
