# Golter

![Logo](https://golter.vercel.app/opengraph-image)

Terminal-based file converter built with Go. It provides a modern, user-friendly Terminal User Interface (TUI) for batch converting images, videos, audio, and documents between various formats.

## Table of Contents

- [Features](#features)
- [Supported Formats](#supported-formats)
  - [Images](#images)
  - [Videos](#videos)
  - [Audio](#audio)
  - [Documents](#documents)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Quick Install](#quick-install)
  - [Build from Source](#build-from-source)
  - [Platform-Specific Setup](#platform-specific-setup)
- [Usage](#usage)
  - [Keyboard Controls](#keyboard-controls)
- [Notes](#notes)
- [License](#license)

## Features

- **Modern TUI Interface:** Beautiful terminal interface with smooth animations and visual feedback.
- **Batch Conversion:** Select multiple files and convert them all at once with concurrent processing.
- **Image Conversion:** Native Go implementation for high-performance image processing with quality control.
- **Video Conversion:** Leverages `ffmpeg` for robust video format support with optimized encoding presets.
- **Audio Conversion:** Convert between various audio formats using `ffmpeg` with bitrate control.
- **Document Conversion:** Support for PDF, Markdown, HTML, EPUB, and doc data conversions (JSON/YAML/XML/TOML/CSV/Excel).
- **Keyboard Navigation:** Full keyboard support with Vim-like keybindings (`j`/`k`, `h`/`l`).
- **Cross-Platform:** Works on Linux, macOS, and Windows.
- **Compression Options:** Choose from High, Balanced, or Compact quality levels.
- **Real-time Progress:** Visual progress indicators during conversion.
- **Smart File Selection:** Only files of the same type can be selected together for consistent conversions.

## Supported Formats

### Images

| Input                            | Output                  |
|----------------------------------|-------------------------|
| `.jpg`, `.jpeg`, `.png`, `.webp` | `.jpg`, `.png`, `.webp` |

**Features:**

- Quality-based compression (92% High, 75% Balanced, 55% Compact)
- Lossless 8-bit WebP with alpha at High quality
- Optimized PNG compression levels

### Videos

| Input                                           | Output                                          |
|-------------------------------------------------|-------------------------------------------------|
| `.mp4`, `.avi`, `.mkv`, `.webm`, `.gif`, `.mov` | `.mp4`, `.avi`, `.mkv`, `.webm`, `.gif`, `.mov` |

**Features:**

- H.264 for MP4 and H.265 for MKV. MKV uses opaque YUV 4:2:0 and pads odd dimensions to even sizes
- VP9 encoding for WebM
- Optimized GIF creation with palette generation
- Multi-threaded encoding
- Fast-start enabled for MP4 streaming

### Audio

| Input                                           | Output                                          |
|-------------------------------------------------|-------------------------------------------------|
| `.mp3`, `.wav`, `.ogg`, `.flac`, `.m4a`, `.aac` | `.mp3`, `.wav`, `.ogg`, `.flac`, `.m4a`, `.aac` |

**Features:**

- Bitrate control (320k High, 192k Balanced, 128k Compact)
- FLAC lossless encoding at the input sample rate
- Vorbis encoding for OGG

> **Note:** Video and audio conversion requires `ffmpeg`. Ebook conversions beyond EPUB require Calibre's `ebook-convert`. DOCX conversions require Pandoc.

### Documents

| Input | Available outputs |
|---|---|
| PDF | Markdown, optimized PDF |
| Markdown | HTML, PDF, DOCX, EPUB, MOBI, AZW3, FB2 |
| HTML | Markdown, DOCX, EPUB, MOBI, AZW3, FB2 |
| DOCX | Markdown, HTML, plain text |
| EPUB, MOBI, AZW, AZW3, FB2 | PDF, HTML, plain text, Markdown, EPUB, MOBI, AZW3, FB2 (excluding the source format) |
| CSV | XLSX, JSON |
| XLSX (first sheet) | CSV, JSON |
| JSON | YAML (.yaml or .yml), XML, CSV, XLSX |
| YAML (.yaml or .yml) | JSON, TOML |
| TOML | YAML (.yaml or .yml) |
| XML | JSON |

AZW is an input format. Use AZW3 for Kindle output. For Excel 97–2003 workbooks, save a copy as XLSX before converting. CSV and XLSX exports to JSON require unique column names. Empty headers are assigned positional names such as `col1`.

**Features:**

- PDF text extraction to Markdown
- Markdown to styled HTML with responsive design
- Markdown/HTML to EPUB conversion
- EPUB/MOBI/AZW/AZW3/FB2 conversions via Calibre (if installed)
- DOCX conversions via Pandoc (if installed)
- PDF compression/optimization
- CSV to Excel conversion with styled headers and auto-fit columns
- Excel to CSV export (exports first sheet)
- JSON ↔ YAML conversion (preserves nested data)
- JSON ↔ XML conversion with automatic root wrapping
- JSON ↔ CSV / Excel (first sheet) conversions
- YAML ↔ TOML conversion

## Installation

### Prerequisites

- **Go 1.24.4+** and a C compiler for CGO (WebP). On Windows, use a GCC toolchain compatible with Go, such as MinGW-w64.
- **ffmpeg** (required for video/audio conversion)
- **Calibre (ebook-convert)** (required for ebook conversions beyond EPUB)
- **Pandoc** (required for DOCX conversions)

### Quick Install

```bash
go install github.com/sametcn99/golter@latest
```

### Build from Source

#### Using Taskfile (Recommended)

```bash
# Install task if not already installed
go install github.com/go-task/task/v3/cmd/task@latest

# Build and run
task build
./bin/golter
```

| Command          | Description                      |
|------------------|----------------------------------|
| `task build`     | Compiles to `bin/golter`         |
| `task run`       | Runs the application             |
| `task clean`     | Removes build artifacts          |
| `task test`      | Runs the test suite              |
| `task fmt`       | Formats the code                 |
| `task build-all` | Builds platform targets (requires matching CGO toolchains) |
| `task install`   | Installs the executable for the current user |

### Platform-Specific Setup

<details>
<summary><b>Linux</b></summary>

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install build-essential ffmpeg calibre pandoc

# Fedora
sudo dnf install gcc ffmpeg calibre pandoc

# Arch Linux
sudo pacman -S base-devel ffmpeg calibre pandoc

# Build
go build -o golter main.go

# Run
./golter
```

</details>

<details>
<summary><b>macOS</b></summary>

```bash
# Install ffmpeg
brew install ffmpeg

# Install Calibre (ebook-convert)
brew install --cask calibre

# Install Pandoc
brew install pandoc

# Build
go build -o golter main.go

# Run
./golter
```

</details>

<details>
<summary><b>Windows</b></summary>

```powershell
# Using winget
winget install ffmpeg

# Calibre (ebook-convert)
winget install calibre.calibre

# Pandoc
winget install --id JohnMacFarlane.Pandoc

# Or using chocolatey
choco install ffmpeg

# Calibre (ebook-convert)
choco install calibre

# Pandoc
choco install pandoc

# Build
go build -o golter.exe main.go

# Run
.\golter.exe
```

</details>

## Usage

Start in current directory:

```bash
./golter
```

Start in a specific directory:

```bash
./golter /path/to/your/media
```

A batch offers only formats shared by every selected file. Compression is available when all inputs support their own format. Existing destinations and inputs that map to the same destination are reported as conflicts. Other files continue. Golter publishes each completed output beside its source. Ctrl+C during a batch cancels pending work and shows the results after running workers finish.

### Keyboard Controls

| Key       | Action                        |
|-----------|-------------------------------|
| `↑` / `k` | Move cursor up                |
| `↓` / `j` | Move cursor down              |
| `←` / `h` | Go to parent directory        |
| `→` / `l` | Enter directory               |
| `Space`   | Select/Deselect file          |
| `a`       | Select all files of same type |
| `d`       | Deselect all files            |
| `Enter`   | Open directory                |
| `c`       | Confirm selection and proceed |
| `/`       | Filter files                  |
| `g`       | Go to top                     |
| `G`       | Go to bottom                  |
| `Esc`     | Go back / Cancel              |
| `q`       | Quit application              |

## Running the tests

Run `go test ./...` for the local suite. The full Linux suite uses FFmpeg and FFprobe, Pandoc, Calibre and Poppler's pdftotext. With these available on PATH, run:

```sh
GOLTER_REQUIRE_TOOLS=1 QT_QPA_PLATFORM=offscreen go test -race -json -count=1 -coverprofile=coverage.out ./... > test-results.jsonl
python3 tests/coverage_summary.py test-results.jsonl coverage.out
```

The strict run fails on missing tools or failed fixture generation. The summary checks every expected media conversion and quality combination. The release workflow also drives the built executable through a native terminal on each platform, converts an XML file and reads back the JSON values.

## Notes

**PATH Note (ebook-convert):**

- **Linux:** Typically available at `/usr/bin/ebook-convert` (or `/snap/bin/ebook-convert`). Ensure the directory is on `PATH`.
- **macOS (Homebrew):** `/opt/homebrew/bin/ebook-convert` (Apple Silicon) or `/usr/local/bin/ebook-convert` (Intel).
- **Windows:** `C:\Program Files\Calibre2\ebook-convert.exe` (or `C:\Program Files (x86)\Calibre2\ebook-convert.exe`). Add the folder to `PATH` if not detected.

**PATH Note (pandoc):**

- **Linux:** Typically available at `/usr/bin/pandoc` or `/usr/local/bin/pandoc`.
- **macOS (Homebrew):** `/opt/homebrew/bin/pandoc` (Apple Silicon) or `/usr/local/bin/pandoc` (Intel).
- **Windows:** `C:\Program Files\Pandoc\pandoc.exe` (or `C:\Program Files (x86)\Pandoc\pandoc.exe`). Add the folder to `PATH` if not detected.

## License

[GPL-3.0](LICENSE)
