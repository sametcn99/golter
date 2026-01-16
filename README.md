# Golter

Terminal-based file converter built with Go. It provides a user-friendly Terminal User Interface (TUI) for batch converting images and videos between various formats.

## Features

- **Terminal User Interface (TUI):** Easy-to-use interface for file navigation and selection.
- **Batch Conversion:** Select multiple files and convert them all at once.
- **Image Conversion:** Native Go implementation for high-performance image processing.
- **Video Conversion:** Leverages `ffmpeg` for robust video format support.
- **Keyboard Navigation:** Vim-like keybindings support (`j`/`k` for navigation).
- **Cross-Platform:** Works on Linux, macOS, and Windows.
- **Compression Options:** Choose quality levels for compression tasks.

## Supported Formats

### Images

- **Input:** `.jpg`, `.jpeg`, `.png`, `.webp`
- **Output:** `.jpg`, `.png`, `.webp`

### Videos

- **Input:** `.mp4`, `.avi`, `.mkv`, `.webm`, `.gif`, `.mov`
- **Output:** `.mp4`, `.avi`, `.mkv`, `.webm`, `.gif`, `.mov`

> **Note:** Video conversion requires `ffmpeg` to be installed on your system.

## Installation & Build

### Prerequisites

- **Go 1.21+** (recommended)
- **ffmpeg** (required for video conversion)

### Build with Taskfile (Recommended)

If you have `task` installed (or install it via `go install github.com/go-task/task/v3/cmd/task@latest`), you can use the following commands to manage the project easily:

| Command          | Description                                  |
|:-----------------|:---------------------------------------------|
| `task build`     | Compiles the application to `bin/golter`     |
| `task run`       | Runs the application directly                |
| `task clean`     | Removes build artifacts                      |
| `task test`      | Runs the test suite                          |
| `task fmt`       | Formats the code                             |
| `task vet`       | Runs static analysis                         |
| `task lint`      | Runs linter (requires `golangci-lint`)       |
| `task deps`      | Downloads and tidies dependencies            |
| `task build-all` | Cross-compiles for Linux, Windows, and macOS |
| `task install`   | Installs binary to system (GOPATH/bin)       |
| `task`           | Shows available commands (default: build)    |

### Linux

1. **Install Dependencies:**

   ```bash
   # Ubuntu/Debian
   sudo apt update && sudo apt install ffmpeg

   # Fedora
   sudo dnf install ffmpeg

   # Arch Linux
   sudo pacman -S ffmpeg
   ```

2. **Build:**

   ```bash
   go build -o golter main.go
   ```

3. **Run:**

   ```bash
   ./golter
   ```

### macOS

1. **Install Dependencies:**

   ```bash
   brew install ffmpeg
   ```

2. **Build:**

   ```bash
   go build -o golter main.go
   ```

3. **Run:**

   ```bash
   ./golter
   ```

### Windows

1. **Install Dependencies:**
   - Download and install [ffmpeg](https://ffmpeg.org/download.html).
   - Add `ffmpeg` to your System PATH.
   - Alternatively, use a package manager:

     ```powershell
     winget install ffmpeg
     # or
     choco install ffmpeg
     ```

2. **Build:**

   ```powershell
   go build -o golter.exe main.go
   ```

3. **Run:**

   ```powershell
   .\golter.exe
   ```

## Usage

Start the application in your current directory (or home directory if not specified):

```bash
./golter
```

Or specify a starting path:

```bash
./golter /path/to/your/media
```

### Controls

| Key            | Action                                             |
|:---------------|:---------------------------------------------------|
| `↑` / `k`      | Move cursor up                                     |
| `↓` / `j`      | Move cursor down                                   |
| `Space`        | Select/Deselect file                               |
| `Enter`        | Open directory                                     |
| `c`            | **Confirm selection** and proceed to format choice |
| `q` / `Ctrl+c` | Quit application                                   |

## How it Works

1. **Navigate:** Use arrow keys to browse your file system.
2. **Select:** Press `Space` to mark files you want to convert. You can select mixed file types (e.g., images and videos).
3. **Confirm:** Press `c` when you are done selecting files.
4. **Choose Action:**
   - **Convert Format:** Change the file format (e.g., JPG to PNG).
   - **Compress Files:** Reduce file size while keeping the same format.
5. **Configure:**
   - If **Convert Format**: Select the target format from the list.
   - If **Compress Files**: Select the desired quality level (High, Medium, Low).
6. **Convert:** The application will process the files. Compressed files will be saved with a `_compressed` suffix.

## License

[GPL-3.0](LICENSE)
