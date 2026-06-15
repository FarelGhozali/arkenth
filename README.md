<div align="center">

# Arkenth

[![Go Report Card](https://goreportcard.com/badge/github.com/FarelGhozali/arkenth)](https://goreportcard.com/report/github.com/FarelGhozali/arkenth)
[![Go Reference](https://pkg.go.dev/badge/github.com/FarelGhozali/arkenth.svg)](https://pkg.go.dev/github.com/FarelGhozali/arkenth)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<br>

*Comprehensive web testing, compiled into a single binary.*

**Arkenth** is an open-source QA automation tool that brings together functional, visual, accessibility, load, and security testing into one dependency-free executable.

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Web Dashboard](#web-dashboard) • [Contributing](#contributing)

</div>

Modern web applications require rigorous testing across multiple disciplines. Arkenth provides a unified solution to help developers, QA engineers, and security researchers ensure software quality without the overhead of managing complex toolchains. Built with Go and Playwright, it features deep security fuzzing, pixel-perfect visual regression, network interception, and an embedded Svelte Web UI. Everything is designed to let you focus on building great software rather than configuring test environments.

## Features

- **Smart Security Fuzzing & API Discovery:** Silently uncovers hidden API routes from frontend JS, forcefully injects payloads (XSS, SQLi) into forms, and tests for Broken Access Control via JWT tampering.
- **Pixel-Perfect Visual Regression:** Takes high-resolution baselines and compares subsequent runs pixel-by-pixel, generating "X-Ray" diffs highlighting UI shifts in bright red.
- **Accessibility (a11y) Engine:** Injects `axe-core` to scan for WCAG violations, actively mutating states (opening hidden UI components like modals and dropdowns) prior to evaluation.
- **Deep Network Interception:** Records failing API calls, fingerprints backend WAF/Servers (Cloudflare, Express, etc.), and captures POST payloads for easier debugging.
- **Concurrent Load Testing:** Bypasses CDN caches by deploying stateful HTTP requests with dynamic payload mutations using highly optimized Goroutines.
- **Embedded Web UI:** Ships with an interactive real-time dashboard built in Svelte, served directly from the single Go binary. No Node.js or extra web servers required.

## Prerequisites

- **Go 1.18** or newer
- **Node.js** (Optional, only if required by internal browser installation scripts)

## Installation

**Option 1: Recommended (Pre-compiled Binary)**
Download the latest pre-compiled executable for your operating system (Windows, macOS, Linux) from the [GitHub Releases](https://github.com/FarelGhozali/arkenth/releases) page. This binary already has the Web UI embedded and requires zero dependencies other than Go.

**Option 2: Build from Source**
Because Arkenth embeds a compiled Svelte frontend, you cannot run a simple `go install`. You must compile the frontend first:

```bash
# 1. Clone the repository
git clone https://github.com/FarelGhozali/arkenth.git
cd arkenth

# 2. Build the Svelte Web Dashboard
cd frontend
npm install
npm run build
cd ..

# 3. Build the Go binary
go build -o arkenth main.go
```

### Install Required Browsers
Since Arkenth relies on Playwright under the hood, you need to download the browser binaries. Run this command once:

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
```

## Quick Start

Arkenth can be operated entirely from your terminal:

```bash
# Scan a target for security, accessibility, and broken links
arkenth scan --target https://example.com --depth 3

# Run in Fast Mode (blocks heavy media and tracking scripts)
arkenth scan --target https://example.com --fast-mode

# Take a Visual Regression baseline
arkenth baseline --target https://example.com/pricing

# Compare the current UI against the baseline
arkenth compare --target https://staging.example.com/pricing
```

## Web Dashboard

Prefer a graphical interface? Launch the embedded Web Dashboard to configure scans, view visual diffs, and read real-time logs right from your browser:

```bash
arkenth ui
```
*Navigate to `http://localhost:8080` (or the indicated port) to access the dashboard.*

## Contributing

We welcome contributions from the community! If you're interested in improving Arkenth, addressing an issue, or adding a new feature, please check out our open issues or submit a Pull Request.

## License

This project is dual-licensed under the [MIT License](LICENSE-MIT) and the [Apache License 2.0](LICENSE-APACHE).
