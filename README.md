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

The application generates two files:

- `<output-report>.json`: A structured JSON representation of the findings.
- `<output-report>.md`: A human-readable Markdown summary.
