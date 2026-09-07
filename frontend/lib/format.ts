export function formatMoney(
  amount: string | number,
  currency?: string | null,
): string {
  const value =
    typeof amount === "number" ? amount : Number.parseFloat(String(amount));
  if (Number.isNaN(value)) {
    return String(amount);
  }

  const code = currency?.trim().toUpperCase();
  if (code && /^[A-Z]{3}$/.test(code)) {
    try {
      return new Intl.NumberFormat(undefined, {
        style: "currency",
        currency: code,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(value);
    } catch {
      // fall through
    }
  }

  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export function formatSignedMoney(
  amount: string | number,
  currency?: string | null,
): string {
  const value =
    typeof amount === "number" ? amount : Number.parseFloat(String(amount));
  if (Number.isNaN(value)) {
    return String(amount);
  }
  const formatted = formatMoney(Math.abs(value), currency);
  if (value > 0) return `+${formatted}`;
  if (value < 0) return `−${formatted}`;
  return formatted;
}

export function moneyTone(amount: string | number): "positive" | "negative" | "neutral" {
  const value =
    typeof amount === "number" ? amount : Number.parseFloat(String(amount));
  if (Number.isNaN(value) || value === 0) return "neutral";
  return value > 0 ? "positive" : "negative";
}

export function currentYearMonthUTC(): string {
  const now = new Date();
  const year = now.getUTCFullYear();
  const month = String(now.getUTCMonth() + 1).padStart(2, "0");
  return `${year}-${month}`;
}

export function formatAccountType(type: string): string {
  return type
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function formatDateLabel(isoDate: string): string {
  if (!isoDate) return "—";
  const date = new Date(isoDate.includes("T") ? isoDate : `${isoDate}T00:00:00Z`);
  if (Number.isNaN(date.getTime())) return isoDate;
  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(date);
}
