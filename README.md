# Golter

![Logo](https://golter.sametcc.me/opengraph-image)

Terminal-based file converter built with Go. It provides a modern, user-friendly Terminal User Interface (TUI) for batch converting images, videos, audio, and documents between various formats.

## Features

- **Modern TUI Interface:** Beautiful terminal interface with smooth animations and visual feedback.
- **Batch Conversion:** Select multiple files and convert them all at once with concurrent processing.
- **Image Conversion:** Native Go implementation for high-performance image processing with quality control.
- **Video Conversion:** Leverages `ffmpeg` for robust video format support with optimized encoding presets.
- **Audio Conversion:** Convert between various audio formats using `ffmpeg` with bitrate control.
- **Document Conversion:** Support for PDF, Markdown, HTML, and EPUB conversions.
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
- WebP lossless mode for highest quality
- Optimized PNG compression levels

### Videos

| Input                                           | Output                                          |
|-------------------------------------------------|-------------------------------------------------|
| `.mp4`, `.avi`, `.mkv`, `.webm`, `.gif`, `.mov` | `.mp4`, `.avi`, `.mkv`, `.webm`, `.gif`, `.mov` |

**Features:**

- H.264/H.265 encoding for MP4/MKV
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
- FLAC lossless support
- Opus/Vorbis encoding for OGG

### Documents

| Input                                     | Output                                                                    |
|-------------------------------------------|---------------------------------------------------------------------------|
| `.pdf`, `.md`, `.html`                    | `.pdf`, `.md`, `.html`, `.epub`, `.mobi`, `.azw`, `.azw3`, `.fb2`         |
| `.epub`, `.mobi`, `.azw`, `.azw3`, `.fb2` | `.epub`, `.mobi`, `.azw`, `.azw3`, `.fb2`, `.pdf`, `.html`, `.txt`, `.md` |
| `.csv`                                    | `.xlsx`                                                                   |
| `.xlsx`, `.xls`                           | `.csv`                                                                    |

**Features:**

- PDF text extraction to Markdown
- Markdown to styled HTML with responsive design
- Markdown/HTML to EPUB conversion
- EPUB/MOBI/AZW/AZW3/FB2 conversions via Calibre (if installed)
- PDF compression/optimization
- CSV to Excel conversion with styled headers and auto-fit columns
- Excel to CSV export (exports first sheet)

> **Note:** Video and audio conversion requires `ffmpeg`. Ebook conversions beyond EPUB require Calibre's `ebook-convert`.

**PATH Note (ebook-convert):**
- **Linux:** Typically available at `/usr/bin/ebook-convert` (or `/snap/bin/ebook-convert`). Ensure the directory is on `PATH`.
- **macOS (Homebrew):** `/opt/homebrew/bin/ebook-convert` (Apple Silicon) or `/usr/local/bin/ebook-convert` (Intel).
- **Windows:** `C:\Program Files\Calibre2\ebook-convert.exe` (or `C:\Program Files (x86)\Calibre2\ebook-convert.exe`). Add the folder to `PATH` if not detected.

## Installation

### Prerequisites

- **Go 1.21+**
- **ffmpeg** (required for video/audio conversion)
- **Calibre (ebook-convert)** (required for ebook conversions beyond EPUB)

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
| `task build-all` | Cross-compiles for all platforms |
| `task install`   | Installs to GOPATH/bin           |
| `task test`      | Runs tests                       |

### Platform-Specific Setup

<details>
<summary><b>Linux</b></summary>

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg calibre

# Fedora
sudo dnf install ffmpeg calibre

# Arch Linux
sudo pacman -S ffmpeg calibre

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

# Or using chocolatey
choco install ffmpeg

# Calibre (ebook-convert)
choco install calibre

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

## License

[GPL-3.0](LICENSE)
