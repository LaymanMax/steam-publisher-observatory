import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PriceTierStackedBar } from "../components/charts/PriceTierStackedBar";
import type { PriceTierItem } from "../types";

describe("PriceTierStackedBar", () => {
  const mockTiers: PriceTierItem[] = [
    {
      tier_name: "Free",
      game_count: 10,
      share: 20.0,
      min_price: 0,
      max_price: 0,
    },
    {
      tier_name: "Low",
      game_count: 20,
      share: 40.0,
      min_price: 5.0,
      max_price: 15.0,
    },
    {
      tier_name: "Premium",
      game_count: 20,
      share: 40.0,
      min_price: 30.0,
      max_price: 50.0,
    },
  ];

  it("renders price tiers and games breakdown", () => {
    render(<PriceTierStackedBar tiers={mockTiers} />);

    expect(screen.getByText("Free")).toBeInTheDocument();
    expect(screen.getByText("Low")).toBeInTheDocument();
    expect(screen.getByText("Premium")).toBeInTheDocument();
    expect(screen.getByText("20.0%")).toBeInTheDocument();
    expect(screen.getAllByText("40.0%").length).toBe(2);
    expect(screen.getByText("(10)")).toBeInTheDocument();
    expect(screen.getAllByText("(20)").length).toBe(2);
  });
});
