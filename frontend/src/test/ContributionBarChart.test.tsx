import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ContributionBarChart } from "../components/charts/ContributionBarChart";
import type { GameOwnershipContribution } from "../types";

describe("ContributionBarChart", () => {
  const mockContributors: GameOwnershipContribution[] = [
    {
      appid: "10",
      name: "Game Alpha",
      owners_lower: 200000,
      owners_upper: 500000,
      owners_midpoint: 350000,
      share_of_midpoint: 50.0,
      cumulative_share: 50.0,
      rank: 1,
    },
    {
      appid: "20",
      name: "Game Beta",
      owners_lower: 100000,
      owners_upper: 200000,
      owners_midpoint: 150000,
      share_of_midpoint: 30.0,
      cumulative_share: 80.0,
      rank: 2,
    },
    {
      appid: "30",
      name: "Game Gamma",
      owners_lower: 50000,
      owners_upper: 100000,
      owners_midpoint: 75000,
      share_of_midpoint: 20.0,
      cumulative_share: 100.0,
      rank: 3,
    },
  ];

  it("renders top contributor items and ranks", () => {
    render(<ContributionBarChart titles={mockContributors} />);

    expect(screen.getByText("Game Alpha")).toBeInTheDocument();
    expect(screen.getByText("Game Beta")).toBeInTheDocument();
    expect(screen.getByText("Game Gamma")).toBeInTheDocument();
    expect(screen.getByText("50.0%")).toBeInTheDocument();
    expect(screen.getByText("30.0%")).toBeInTheDocument();
    expect(screen.getByText("20.0%")).toBeInTheDocument();
  });

  it("renders fallback when no titles provided", () => {
    render(<ContributionBarChart titles={[]} />);
    expect(
      screen.getByText("No valid ownership contributions to chart."),
    ).toBeInTheDocument();
  });
});
