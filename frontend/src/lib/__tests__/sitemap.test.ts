import { generateSitemapEntries } from "@/lib/sitemap";

describe("generateSitemapEntries", () => {
  it("TestGenerateSitemapEntries_ReturnsArray", () => {
    const result = generateSitemapEntries();

    expect(Array.isArray(result)).toBe(true);
    expect(result.length).toBeGreaterThan(0);
  });

  it("TestGenerateSitemapEntries_ValidURLs", () => {
    const result = generateSitemapEntries();

    for (const entry of result) {
      expect(entry.url.startsWith("/")).toBe(true);
    }
  });

  it("TestGenerateSitemapEntries_ValidPriorities", () => {
    const result = generateSitemapEntries();

    for (const entry of result) {
      expect(entry.priority).toBeGreaterThanOrEqual(0);
      expect(entry.priority).toBeLessThanOrEqual(1);
    }
  });

  it("TestGenerateSitemapEntries_HasRequiredFields", () => {
    const result = generateSitemapEntries();

    for (const entry of result) {
      expect(entry).toHaveProperty("url");
      expect(entry).toHaveProperty("lastmod");
      expect(entry).toHaveProperty("changefreq");
      expect(entry).toHaveProperty("priority");
    }
  });
});
