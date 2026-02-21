# Enterprise Web QA Automation CLI

A comprehensive, automated Web QA Testing CLI software built with Go, `spf13/cobra`, and `playwright-go`.

This tool acts as a dynamic scanner and tester for any given target URL, providing modular components for crawling, deep network interception, active fuzzing, and pixel-perfect visual regression testing.

## 🚀 Features

1. **Modular CLI Interface (`cobra`)**
   - Execute specialized subcommands (`scan`, `baseline`, `compare`) with robust flags instead of a single script.
2. **Deep Network Interceptor**
   - Intercepts requests to find failed HTTP statuses (4xx/5xx).
   - Speed up scans using `--fast-mode` which forcibly blocks tracking scripts, heavy images, and fonts.
3. **Advanced Active QA (Fuzzing & Rage Clicks)**
   - Automatically detects forms, text areas, and inputs. Injects predefined payloads (XSS, SQLi, boundary values).
   - Simulates aggressive DOM interaction ("Rage clicks") on buttons and captures any resulting Javascript crashes or exceptions.
4. **Visual Regression Testing (Pixel-Perfect)**
   - Take pristine baseline screenshots of your app's layout.
   - Run comparison scans later to generate highlighted markdown diff reports revealing unauthorized CSS layout shifts.
5. **Mobile Emulation & State Management**
   - Test responsive designs natively using Playwright's device profiles (e.g., `--mobile-emulation="iPhone 13"`).
   - Inject saved authentication cookies to test gated dashboards (`--auth-json`).

## ⚙️ Installation

### Prerequisites

- Go 1.18+

### Setup Instructions

1. **Clone the repository**:

   ```bash
   git clone <repo-url>
   cd web-qa-automation
   ```

2. **Install Dependencies**:

   ```bash
   go mod tidy
   ```

3. **Install Playwright Browsers**:

   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
   ```

4. **Build the CLI**:
   ```bash
   go build -o web-qa
   ```

## 📖 Usage Guide

Run the compiled executable with the required commands and flags:

### 1. The Standard QA Scan

Executes aggressive fuzzing, network interception, and generates `qa_audit_report.md`:

```bash
./web-qa scan --target="https://example.com" --depth=2 --fast-mode
```

### 2. Authenticated Scans (Dashboards)

Pass a Playwright storage state JSON to test logged-in areas:

```bash
./web-qa scan --target="https://example.com/admin" --auth-json="auth.json"
```

### 3. Visual Regression: Baseline Capture

Take canonical layout snapshots (bypasses fuzzing to ensure pages are captured cleanly):

```bash
./web-qa baseline --target="https://example.com" --depth=2
```

_Screenshots are saved mapped to `proofs/baseline/`._

### 4. Visual Regression: Compare Phase

Take current layout snapshots and compare them against the baseline pixel-by-pixel:

```bash
./web-qa compare --target="https://example.com" --depth=2
```

_Outputs a dedicated `visual_regression_report.md` alongside highlighting differences in `proofs/diff/`._

### Global Flags

- `--target` : (Required) The root URL you want to crawl and test.
- `--depth` : (Optional) The maximum link-following depth. Default is 1.
- `--fast-mode`: (Optional) Block media/trackers to boost scan speed.
- `--mobile-emulation`: (Optional) Device string, e.g., 'iPhone 13'.
- `--record-video`: (Optional) Saves WebM playbacks of the test session.
- `--auth-json`: (Optional) Path to a stored session state for authentication.

---

## 🛡️ Disclaimer

This framework executes highly aggressive DOM traversal and form fuzzing techniques. It is intended for **authorized QA testing and security auditing only**. Do not use it against production applications without proper isolation, as active fuzzing will manipulate databases and may trigger security alerts.
