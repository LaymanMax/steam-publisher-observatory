# Metrics & Analytical Methodology

## 1. Important Analytical Disclaimer

> **Crucial Distinction**: All metrics derived in Steam Publisher Observatory reflect **estimated catalog ownership ranges** and public store metadata collected from Steam. They **do NOT represent audited commercial revenue, net publisher receipts, retail units sold, profit margins, or distinct concurrent account players**.

Steam does not publish private financial statements or exact player counts. The estimated ownership ranges (e.g. `20,000 - 50,000` or `1,000,000 - 2,000,000`) provide indicative, relative scale comparisons across portfolios.

---

## 2. Portfolio Dependency & Concentration

### 2.1 Top-N Concentration Shares

Let a publisher's catalog have $N$ titles with valid estimated ownership midpoints $m_1, m_2, \dots, m_N$, sorted in descending order ($m_1 \ge m_2 \ge \dots \ge m_N$).

The total catalog estimated midpoint is:
$$M = \sum_{i=1}^N m_i$$

For any subset of top $k$ titles ($k \in \{1, 3, 5, 10\}$), the concentrated ownership share is:
$$S_k = \frac{\sum_{i=1}^k m_i}{M} \times 100\%$$

_Example_: If a publisher has 50 titles with a combined midpoint of 10,000,000 owners, and its 5 largest titles contribute 7,200,000 owners, the **Top 5 Share is exactly 72.0%**.

### 2.2 Herfindahl-Hirschman Index (HHI)

To capture concentration across the entire distribution rather than an arbitrary cutoff, the application computes the normalized market HHI.

Let title $i$'s percentage share of the catalog midpoint be:
$$s_i = \frac{m_i}{M} \times 100$$

The portfolio HHI is calculated as:
$$\text{HHI} = \sum_{i=1}^N s_i^2$$

**Classification Scale**:

- **Unconcentrated / High Diversity** ($\text{HHI} < 1,500$): The portfolio's audience is widely distributed across many titles.
- **Moderate Concentration** ($1,500 \le \text{HHI} < 2,500$): Noticeable reliance on a primary cluster of leading titles.
- **High Concentration / Monopolistic Dependence** ($\text{HHI} \ge 2,500$): Severe catalog risk where 1 or 2 hit games dictate the vast majority of portfolio reach.

### 2.3 Effective Title Count ($N_{\text{eff}}$)

The inverse of the unscaled fractional HHI indicates the equivalent number of equal-sized titles that would yield the same level of concentration:
$$N_{\text{eff}} = \frac{1}{\sum_{i=1}^N \left(\frac{m_i}{M}\right)^2} = \frac{10,000}{\text{HHI}}$$

_Interpretation_: If a publisher has 200 catalog games but an effective title count of $2.8$, its portfolio audience footprint is mathematically equivalent to having just 3 equal-sized titles.

### 2.4 Lorenz Curve & Gini Coefficient

The application computes 10-point cumulative decile coordinates of the Lorenz curve comparing cumulative percentage of titles against cumulative percentage of estimated owners, visually illustrating deviation from the line of perfect equality ($45^\circ$).

---

## 3. Multi-Genre Proportional Allocation

Steam titles frequently tag multiple genres (e.g. `Action`, `Adventure`, `RPG`). Summing raw counts across multi-genre titles distorts the portfolio total (summing to 200% or 300% of the catalog).

### 3.1 Proportional 1/K Weighting

If a title $g$ is tagged with $K_g$ distinct genres, each genre receives a proportional weight of:
$$w_{g, j} = \frac{1}{K_g}$$

For any genre $j$, its weighted title credit is:
$$W_j = \sum_{g \in \text{Catalog}, j \in \text{Genres}(g)} \frac{1}{K_g}$$

Because $\sum_{j \in \text{Genres}(g)} w_{g, j} = 1.0$ for every title, the sum of weighted genre credits across all genres equals the exact number of genre-classified catalog games:
$$\sum_j W_j = N_{\text{classified}}$$

The **Weighted Genre Share** is:
$$\text{Share}_j = \frac{W_j}{N_{\text{classified}}} \times 100\%$$

This guarantees the genre composition pie/bar chart sums to **exactly 100.0%**.

---

## 4. Tukey IQR Outlier Detection

To detect portfolio outliers (breakout hits or extreme underperformers) without being misled by non-normal skew, the system uses non-parametric **Tukey Interquartile Range (IQR) fences**:

1. Order title values $x_1 \le x_2 \le \dots \le x_N$.
2. Compute first quartile $Q_1$ (25th percentile) and third quartile $Q_3$ (75th percentile).
3. Compute interquartile range:
   $$\text{IQR} = Q_3 - Q_1$$
4. Define outlier fences:
   $$\text{Lower Fence} = Q_1 - 1.5 \times \text{IQR}$$
   $$\text{Upper Fence} = Q_3 + 1.5 \times \text{IQR}$$

Titles with metrics exceeding the Upper Fence are flagged as **High Outliers** (breakout hits). Titles below the Lower Fence are flagged as **Low Outliers**.

---

## 5. Pricing Tiers

Catalog prices are segmented into commercial tiers reflecting standard Steam store monetization categories:

- **Free**: $\$0.00$ (Free-to-play, demos, promotional)
- **Very Low**: $\$0.01 - \$4.99$ (Micro-indies, legacy discounts)
- **Low**: $\$5.00 - \$14.99$ (Standard indie price point)
- **Mid-range**: $\$15.00 - \$29.99$ (AA titles, premium indies)
- **Premium**: $\$30.00 - \$49.99$ (Major AA releases, specialty titles)
- **High-priced**: $\$50.00+$ (Full AAA releases)
- **Missing**: Unpriced or unlisted titles

---

## 6. Data Quality Completeness Scoring

Each title receives an objective completeness score $C \in [0, 100]$ based on metadata availability:

- **Core Title & AppID**: 20 pts
- **Publisher Specification**: 20 pts
- **Developer Specification**: 15 pts
- **Valid Ownership Range**: 15 pts
- **Genre Classification**: 10 pts
- **Store Description**: 10 pts
- **Release Date**: 5 pts
- **Header Asset Image**: 5 pts

Titles with scores $< 40$ are marked as `is_sparse`. Records containing keywords `playtest`, `closed beta`, or `internal test` with 0 owners are classified as `is_playtest` to allow analytical filtering.
