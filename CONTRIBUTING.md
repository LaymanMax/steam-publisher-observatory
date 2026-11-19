# Contributing to Steam Publisher Observatory

We welcome contributions of all kinds to improve this project! Here is how you can get involved:

## Workflow

1. **Fork the Repository**: Start by forking the repository to your GitHub account.

2. **Clone Your Fork**: Clone the forked repository to your local machine:

   ```bash
   git clone https://github.com/<your-username>/steam-publisher-observatory.git
   cd steam-publisher-observatory
   ```

3. **Create a Feature Branch**:

   ```bash
   git checkout -b feature/my-new-feature
   ```

4. **Make Changes**: Follow existing Go and TypeScript code styles and conventions.

5. **Run Tests**:
   - Backend tests:
     ```bash
     cd backend && go test -v ./...
     ```
   - Frontend tests:
     ```bash
     cd frontend && pnpm test
     ```
   - Verify production bundle:
     ```bash
     cd frontend && pnpm build
     ```

6. **Commit Changes**: Use clear and descriptive commit messages:

   ```bash
   git commit -m "feat(analytics): add support for custom price tier thresholds"
   ```

7. **Push to Your Fork**:

   ```bash
   git push origin feature/my-new-feature
   ```

8. **Submit a Pull Request**: Open a pull request from your branch against the `main` branch. Provide a detailed summary of your changes, motivation, and test verification.

---

## Guidelines

- Follow standard Go formatting (`gofmt`, `go vet`).
- Ensure all TypeScript types are strict and clean (`tsc`).
- Include unit/integration tests for any new analytical metric or API endpoint.
- Do not commit large Parquet files, database files (`.duckdb`), or credentials.

## Reporting Issues

Feel free to report bugs or suggest enhancements by creating an issue. Include detailed information about the problem, steps to reproduce it, and the environment (OS, Go version, Node version).
