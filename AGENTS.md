# Flow — AI Agent Guidelines

A native desktop application for GNOME that interfaces with the local HomeWizard API to monitor real-time electricity, water, and gas consumption.

## Environment & Safety Rules

> **Nix Flake Requirement**
> You must run commands inside the project's **Nix flake development shell** (`nix develop`).
> If required CLI tools (such as `go`) are not available in your shell environment, **stop immediately and notify the user** to enter the Nix shell.

## Core Tech Stack

* **Language:** Go
* **GUI Framework:** Libadwaita
* **API:** HomeWizard Local API (JSON / Local REST endpoints)

## Code Quality & Architectural Principles

* **Function Granularity:** Prefer short, single-responsibility, clean functions over monolithic blocks.
* **Separation of Concerns:**
  * Keep the **HomeWizard API integration** (HTTP/networking/parsing) distinct from the **UI rendering**.
  * Keep domain data models agnostic of presentation logic.
* **HomeWizard Module:** Use the go-homewizard Go module for HomeWizard integration.
* **Non-Blocking UI:** Handle local network requests asynchronously to avoid freezing the main thread.
* **GNOME Integration:** Follow GNOME Human Interface Guidelines (HIG) for layout, iconography, and styling.
* **HomeWizard API version:** Use the local v1 API, NOT v2.

## Common Commands

Run these inside the Nix environment:

```bash
# Build the application
go build -v ./...

# Run tests
go test ./...

# Format code
go fmt ./...
```
