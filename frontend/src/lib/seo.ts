import type { Metadata } from "next";

type PlainObject = Record<string, unknown>;

function isPlainObject(value: unknown): value is PlainObject {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function cloneValue<T>(value: T): T {
  if (Array.isArray(value)) {
    return value.map((item) => cloneValue(item)) as T;
  }

  if (isPlainObject(value)) {
    const cloned: PlainObject = {};
    for (const [key, item] of Object.entries(value)) {
      cloned[key] = cloneValue(item);
    }
    return cloned as T;
  }

  return value;
}

function deepMerge<T>(base: T, overrides?: Partial<T>): T {
  if (overrides === undefined) {
    return cloneValue(base);
  }

  const baseValue = cloneValue(base);
  if (!isPlainObject(baseValue) || !isPlainObject(overrides)) {
    return cloneValue(overrides as T);
  }

  const merged: PlainObject = { ...baseValue };

  for (const [key, overrideValue] of Object.entries(overrides)) {
    if (overrideValue === undefined) {
      continue;
    }

    const currentValue = merged[key];
    if (isPlainObject(currentValue) && isPlainObject(overrideValue)) {
      merged[key] = deepMerge(currentValue, overrideValue);
      continue;
    }

    merged[key] = cloneValue(overrideValue);
  }

  return merged as T;
}

export const DEFAULT_METADATA: Metadata = {
  title: "Create Go App",
  description: "A modern full-stack starter powered by Go and Next.js.",
  openGraph: {
    title: "Create Go App",
    description: "A modern full-stack starter powered by Go and Next.js.",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Create Go App",
    description: "A modern full-stack starter powered by Go and Next.js.",
  },
};

export function buildMetadata(overrides?: Metadata): Metadata {
  return deepMerge(DEFAULT_METADATA, overrides);
}
