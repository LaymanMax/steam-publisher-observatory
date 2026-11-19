import { describe, expect, it } from "vitest";
import { api } from "../services/api";

describe("API service client", () => {
  it("defines the required endpoints for publisher analytics", () => {
    expect(typeof api.getOverview).toBe("function");
    expect(typeof api.getDatasetStatus).toBe("function");
    expect(typeof api.listPublishers).toBe("function");
    expect(typeof api.getPublisherSummary).toBe("function");
    expect(typeof api.getPublisherCatalog).toBe("function");
    expect(typeof api.getPublisherDependency).toBe("function");
    expect(typeof api.getPublisherGenres).toBe("function");
    expect(typeof api.getPublisherPrices).toBe("function");
    expect(typeof api.getPublisherDevelopers).toBe("function");
    expect(typeof api.getPublisherTimeline).toBe("function");
    expect(typeof api.getPublisherOutliers).toBe("function");
    expect(typeof api.getPublisherQuality).toBe("function");
    expect(typeof api.comparePublishers).toBe("function");
    expect(typeof api.getDatasetQuality).toBe("function");
    expect(typeof api.getDatasetIssues).toBe("function");
    expect(typeof api.listDevelopers).toBe("function");
    expect(typeof api.listAliases).toBe("function");
    expect(typeof api.createAlias).toBe("function");
    expect(typeof api.deleteAlias).toBe("function");
  });
});
