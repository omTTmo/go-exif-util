# AI Prompt EXIF Reader

A simple desktop application built with Go and Fyne that lets you browse a folder of images, preview them, and extract embedded AI-prompt metadata (e.g., Stable Diffusion "UserComment" payloads) via ExifTool.
## Features

- **Folder Browser**
  Pick any folder and display only its image files (`.jpg`, `.jpeg`, `.png`, `.webp`) in a sidebar.
- **Live Preview**
  Preview the image and its EXIF prompt update instantly.
- **EXIF Prompt Extraction**
  Reads `UserComment`, `ImageDescription`, or `XPComment` and pretty-prints JSON if present.

## Prerequisites

1. **Go 1.18+** installed.
2. **ExifTool CLI** installed and reachable—GUI apps use a minimal `$PATH`. Ensure it's in one of:
   - `/opt/homebrew/bin/exiftool` (Apple Silicon)
   - `/usr/local/bin/exiftool` (Intel macOS / Linux Homebrew)
   - or globally on your `PATH`

## Installation & Setup

```bash
git clone git@github.com:omTTmo/go-exif-util.git
cd aiprompt-exif-reader
go mod tidy
```

## Running locally
go build -o aiprompt-reader
./aiprompt-reader


