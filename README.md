# Web QA Automation CLI

A comprehensive, automated Web QA Testing CLI software written in Go using the `playwright-go` library.

This tool acts as a dynamic scanner and tester for any given target URL, providing modular components for crawling, passive network/console monitoring, and active fuzzing.

## Features

1. **CLI Interface & Configuration**
   - Setup scan parameters like target URL, scanning depth, and report file generation.
2. **Spider/Crawler Module**
   - Navigates through internal links up to the specified depth.
3. **Passive QA (Network & Console Monitor)**
   - Intercepts requests to find failed HTTP statuses (4xx/5xx).
   - Listens to browser console logs and captures JavaScript errors / warnings.
4. **Active QA (Fuzzing/Edge Cases)**
   - Automatically detects forms and injects predefined payloads (XSS, SQLi, emojis, long strings).
   - Captures any crashes or uncaught exceptions during form submission.
5. **Report Generator**
   - Outputs findings in clean JSON and readable Markdown formats.

## Installation

### Prerequisites

- Go 1.18+
- Node.js (Optional, but sometimes required implicitly for driver downloads)

### Setup Instructions

1. **Clone the repository** (if applicable) or navigate to the source directory:

   ```bash
   cd web-qa-automation
   ```

2. **Initialize the Go Module** (if not already initialized):

   ```bash
   go mod init web-qa-automation
   ```

3. **Install Dependencies**:
   This will download the required Playwright-Go library.

   ```bash
   go get github.com/playwright-community/playwright-go
   ```

4. **Install Playwright Browsers**:
   The code automatically attempts to install Playwright drivers and browsers upon running. However, you can also install them manually:

   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
   ```

5. **Build the CLI**:
   ```bash
   go build -o web-qa
   ```

## Usage

Run the compiled executable with the required flags:

```bash
./web-qa --target-url="https://example.com" --depth=2 --output-report="scan_results"
```

### CLI Flags

- `--target-url` : (Required) The root URL you want to crawl and test.
- `--depth` : (Optional) The maximum link-following depth. Default is 1.
- `--output-report` : (Optional) Prefix for the generated report files. Default is `report`.

### Output

The application generates two files by default in your current directory:

#### 1. JSON Report (`<output-report>.json`)

A structured JSON file ideal for parsing by other tools or CI/CD pipelines. It contains:

- `target_url`: The initial URL scanned.
- `scanned_urls`: A complete list of all internal pages visited during the crawl.
- `network_findings`: Array of failed HTTP requests (status >= 400).
- `js_findings`: Array of JavaScript console errors and warnings.
- `vulnerabilities`: Array of captured crashes or unhandled exceptions from the active fuzzing module.

#### 2. Markdown Report (`<output-report>.md`)

A human-readable summary that is perfectly formatted for reading or sharing with the team. It groups findings into clear sections:

- **Scanned URLs**: Lists all successfully crawled paths.
- **Network Findings**: Highlights broken links or failed API calls.
- **JS Console Findings**: Shows client-side scripting errors.
- **Vulnerabilities**: Details the exact payload and form that caused a crash or unexpected error.

---

## 🏗️ Architecture & Modules

The tool is built with modularity in mind, making it easy to extend or customize:

- **`config`**: Handles parsing of CLI arguments and sets up the scanning parameters.
- **`crawler`**: A BFS (Breadth-First Search) spider that navigates the target application. It extracts `href` links from the DOM and recursively follows internal links up to the specified `--depth`, avoiding infinite loops by tracking visited URLs.
- **`tester`**: The core QA engine containing two sub-modules:
  - **Passive Monitors**: Hooks into Playwright's page events (`OnConsole`, `OnResponse`, `OnRequestFailed`) to silently observe the site's behavior during normal navigation.
  - **Active Fuzzer**: Locates all `<form>` tags and input fields. It injects a suite of predefined payloads (e.g., `-1`, emojis, max length strings, basic XSS `<script>alert(1)</script>`, and SQLi `' OR 1=1 --`) and attempts to submit the forms, monitoring the page for unhandled exceptions or crashes.
- **`reporter`**: Aggregates all findings into structured structs and manages the robust export to JSON and Markdown.

---

## 💡 Advanced Usage Examples

### 1. Shallow Scan (Quick Check)

Run a quick test on the homepage without following any links (Depth 0).

```bash
./web-qa --target-url="https://example.com" --depth=0 --output-report="quick_check"
```

### 2. Deep Application Scan

Crawl deep into the application (e.g., Depth 3) to test internal pages and forms. _Note: Deep scans can take significantly longer._

```bash
./web-qa --target-url="https://example.com" --depth=3 --output-report="deep_scan"
```

---

## 🛡️ Disclaimer

This tool is intended for **authorized QA testing and security auditing only**. Do not use it against applications or websites you do not own or do not have explicit permission to test, as active fuzzing may cause data corruption or trigger security alerts.
