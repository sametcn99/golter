"""Exercise the release executable in a native pseudoterminal."""
import json
import os
from pathlib import Path
from queue import Empty, Queue
import sys
from tempfile import TemporaryDirectory
from threading import Thread
from time import monotonic

if os.name == "nt":
    from winpty import PtyProcess as Terminal
else:
    from ptyprocess import PtyProcessUnicode as Terminal


def smoke(binary, directory):
    (directory / "sample.xml").write_text("<root><name>Golter</name></root>", encoding="utf8")
    terminal = Terminal.spawn([str(binary)], cwd=str(directory),
                              env=dict(os.environ, TERM="xterm-256color"), dimensions=(36, 120))
    chunks = Queue()

    def read_output():
        try:
            while True:
                chunk = terminal.read(4096)
                if not chunk:
                    break
                chunks.put(chunk)
        except EOFError:
            pass
        finally:
            chunks.put(None)

    Thread(target=read_output, daemon=True).start()
    pending = ""

    def expect(marker):
        nonlocal pending
        deadline = monotonic() + 20
        while marker not in pending:
            try:
                chunk = chunks.get(timeout=max(0, deadline - monotonic()))
            except Empty:
                raise AssertionError(f"Timed out waiting for {marker!r}: {pending[-1000:]!r}")
            if chunk is None:
                raise AssertionError(f"Early terminal exit waiting for {marker!r}: {pending[-1000:]!r}")
            pending += chunk
        pending = pending.split(marker, 1)[1]

    try:
        # Answer the terminal background-color query emitted by Lip Gloss on Unix.
        if os.name != "nt":
            terminal.write("\x1b]11;rgb:0000/0000/0000\x1b\\")
        expect("Select files to convert")
        terminal.write("a")
        expect("1 documents selected")
        terminal.write("c")
        expect("Choose action")
        terminal.write("\r")
        expect("Select target format")
        terminal.write("\r")
        expect("Successfully converted 1")
        result = json.loads((directory / "sample.json").read_text(encoding="utf8"))
        assert result == {"root": {"name": "Golter"}}, result
        terminal.write("\x1b")
        expect("Select files to convert")
        terminal.write("q")
        expect("Quit Confirmation")
        terminal.write("y")
        deadline = monotonic() + 20
        while chunks.get(timeout=max(0, deadline - monotonic())) is not None:
            pass
        assert terminal.wait() == 0
    finally:
        if terminal.isalive():
            terminal.terminate(force=True)
        terminal.close(force=True)


if __name__ == "__main__":
    with TemporaryDirectory() as workdir:
        smoke(Path(sys.argv[1]).resolve(), Path(workdir))
