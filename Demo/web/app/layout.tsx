import type { Metadata } from "next";
import "@fontsource-variable/montserrat";
import "./globals.css";

export const metadata: Metadata = {
  title: "Vantaca × Northwind Demo",
  description: "Interview demonstration of a safety-conscious financial integration",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
