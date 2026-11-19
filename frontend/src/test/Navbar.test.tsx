import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { Navbar } from "../components/layout/Navbar";

describe("Navbar", () => {
  it("renders branding and primary navigation links", () => {
    render(
      <MemoryRouter>
        <Navbar />
      </MemoryRouter>,
    );

    expect(screen.getByText("Steam Publisher")).toBeInTheDocument();
    expect(screen.getAllByText("Observatory").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Publishers Directory")).toBeInTheDocument();
    expect(screen.getByText("Compare")).toBeInTheDocument();
    expect(screen.getByText("Data Quality")).toBeInTheDocument();
    expect(screen.getByText("Dataset Manager")).toBeInTheDocument();
  });
});
