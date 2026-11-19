import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatCard } from "../components/common/StatCard";

describe("StatCard", () => {
  it("renders label, value, and subtext correctly", () => {
    render(
      <StatCard
        label="Active Publishers"
        value="14,200"
        subtext="Across catalog"
      />,
    );

    expect(screen.getByText("Active Publishers")).toBeInTheDocument();
    expect(screen.getByText("14,200")).toBeInTheDocument();
    expect(screen.getByText("Across catalog")).toBeInTheDocument();
  });

  it("renders badge when provided", () => {
    render(
      <StatCard
        label="Concentration Index"
        value="0.3200"
        badge={{
          text: "Moderate",
          variant: "warning",
        }}
      />,
    );

    expect(screen.getByText("Concentration Index")).toBeInTheDocument();
    expect(screen.getByText("0.3200")).toBeInTheDocument();
    expect(screen.getByText("Moderate")).toBeInTheDocument();
  });
});
