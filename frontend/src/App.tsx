import React from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { Navbar } from "./components/layout/Navbar";
import { PublisherCatalogPage } from "./features/catalog/PublisherCatalogPage";
import { ComparisonPage } from "./features/comparison/ComparisonPage";
import { DashboardPage } from "./features/dashboard/DashboardPage";
import { DataQualityPage } from "./features/data-quality/DataQualityPage";
import { DatasetManagerPage } from "./features/dataset/DatasetManagerPage";
import { PublisherReportPage } from "./features/portfolio/PublisherReportPage";
import { PublisherDirectoryPage } from "./features/publishers/PublisherDirectoryPage";

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-[#090d16] text-[#e2e8f0] flex flex-col font-sans">
        <Navbar />

        <main className="flex-1">
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/publishers" element={<PublisherDirectoryPage />} />
            <Route
              path="/publishers/:publisherId"
              element={<PublisherReportPage />}
            />
            <Route
              path="/publishers/:publisherId/catalog"
              element={<PublisherCatalogPage />}
            />
            <Route path="/compare" element={<ComparisonPage />} />
            <Route path="/data-quality" element={<DataQualityPage />} />
            <Route path="/dataset" element={<DatasetManagerPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>

        <footer className="border-t border-[#1e293b] bg-[#070b12] py-6 text-center text-xs text-slate-500 font-mono">
          <div className="container-max flex flex-col sm:flex-row items-center justify-between gap-2">
            <span>
              Steam Publisher Observatory • Portfolio Risk & Concentration
              Analytics
            </span>
            <span>Data source: FronkonGames/steam-games-dataset</span>
          </div>
        </footer>
      </div>
    </BrowserRouter>
  );
};
export default App;
