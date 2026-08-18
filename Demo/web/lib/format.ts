const moneyPattern = /^(-?)(\d+)\.(\d{2})$/;

export function formatMoney(value: string, currency: string): string {
  const match = moneyPattern.exec(value);
  if (!match) {
    return `${currency} ${value}`;
  }
  const [, sign, whole, fraction] = match;
  const grouped = whole.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  const symbol = currency === "USD" ? "$" : `${currency} `;
  return `${sign}${symbol}${grouped}.${fraction}`;
}

export function formatTimestamp(value: string): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
}

export type TransferTone = "neutral" | "info" | "success" | "danger" | "warning";

export function transferTone(status: string): TransferTone {
  switch (status) {
    case "PENDING":
      return "info";
    case "POSTED":
      return "success";
    case "FAILED":
    case "RETURNED":
      return "danger";
    case "UNKNOWN":
      return "warning";
    default:
      return "neutral";
  }
}
