# Flow — AI Agent Guidelines

A native GTK desktop application for GNOME that interfaces with the local HomeWizard API to monitor real-time electricity, water, and gas consumption.

---

## Environment & Safety Rules

> **Nix Flake Requirement**
> You must run commands inside the project's **Nix flake development shell** (`nix develop`).
> If required CLI tools (such as `go` or GTK dependencies) are not available in your shell environment, **stop immediately and notify the user** to enter the Nix shell.

---

## Core Tech Stack

* **Language:** Go
* **GUI Framework:** GTK / Libadwaita
* **API:** HomeWizard Local API (JSON / Local REST endpoints)

---

## Code Quality & Architectural Principles

* **Function Granularity:** Prefer short, single-responsibility, clean functions over monolithic blocks.
* **Separation of Concerns:**
  * Keep the **HomeWizard API integration** (HTTP/networking/parsing) distinct from **GTK UI rendering**.
  * Keep domain data models agnostic of presentation logic.
* **Non-Blocking UI:** Handle local network requests asynchronously to avoid freezing the GTK main thread.
* **GNOME Integration:** Follow GNOME Human Interface Guidelines (HIG) for layout, iconography, and styling.

---

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
