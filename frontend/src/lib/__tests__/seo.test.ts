import { DEFAULT_METADATA, buildMetadata } from "@/lib/seo";

describe("buildMetadata", () => {
  it("TestBuildMetadata_Defaults", () => {
    const metadata = buildMetadata();

    expect(metadata).toHaveProperty("title");
    expect(metadata).toHaveProperty("description");
    expect(metadata).toHaveProperty("openGraph");
    expect(metadata).toHaveProperty("twitter");

    expect(metadata.title).toEqual(DEFAULT_METADATA.title);
    expect(metadata.description).toEqual(DEFAULT_METADATA.description);
    expect(metadata.openGraph).toEqual(DEFAULT_METADATA.openGraph);
    expect(metadata.twitter).toEqual(DEFAULT_METADATA.twitter);
  });

  it("TestBuildMetadata_Override", () => {
    const metadata = buildMetadata({ title: "X" });

    expect(metadata.title).toBe("X");
    expect(metadata.description).toEqual(DEFAULT_METADATA.description);
    expect(metadata.openGraph).toEqual(DEFAULT_METADATA.openGraph);
    expect(metadata.twitter).toEqual(DEFAULT_METADATA.twitter);
  });

  it("TestBuildMetadata_NoMutation", () => {
    const first = buildMetadata({ openGraph: { title: "First" } });
    const second = buildMetadata({ openGraph: { title: "Second" } });
    const defaults = buildMetadata();

    expect(first.openGraph?.title).toBe("First");
    expect(second.openGraph?.title).toBe("Second");
    expect(defaults.openGraph?.title).toEqual(DEFAULT_METADATA.openGraph?.title);
  });
});
