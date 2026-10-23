# Flow

A native desktop application for GNOME that interfaces with the local [HomeWizard](https://www.homewizard.com/) API to monitor real-time electricity, water, and gas consumption.

> [!NOTE]
> This is a personal project and may not work for everyone. Use at your own risk.

## Running

Run the app:

```
nix run github:cimm/flow
```

## Compiling

Build the app:

```bash
nix build
```

This produces `./result/bin/flow`, which is a properly wrapped binary that will work on NixOS — even after a restart or garbage collection.

## AI Disclaimer

his application was developed with the help of a Large Language Model (LLM). As a result, the code may be slop.
