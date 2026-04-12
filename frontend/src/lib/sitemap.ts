export type SitemapEntry = {
  url: string;
  lastmod: string;
  changefreq: "always" | "hourly" | "daily" | "weekly" | "monthly" | "yearly" | "never";
  priority: number;
};

export function generateSitemapEntries(): SitemapEntry[] {
  return [
    {
      url: "/",
      lastmod: "2026-01-01T00:00:00Z",
      changefreq: "daily",
      priority: 1,
    },
    {
      url: "/login",
      lastmod: "2026-01-01T00:00:00Z",
      changefreq: "weekly",
      priority: 0.8,
    },
    {
      url: "/register",
      lastmod: "2026-01-01T00:00:00Z",
      changefreq: "weekly",
      priority: 0.8,
    },
  ];
}
