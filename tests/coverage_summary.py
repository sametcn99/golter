"""Check the full-tool Linux matrices and write their results to the CI summary."""
import json
from pathlib import Path
import sys

events, profile = map(Path, sys.argv[1:])
passed = set()
with events.open() as stream:
    for line in stream:
        event = json.loads(line)
        if event.get("Action") == "skip":
            raise SystemExit(f"Unexpected skip in full integration run: {event.get('Test', event['Package'])}")
        if event.get("Action") == "pass" and "Test" in event:
            passed.add(event["Test"])

print("### Conversion matrices\n")
print("| Suite | Expected | Passed |")
print("|---|---:|---:|")
for kind, formats in {
        "Audio": "mp3 wav ogg flac m4a aac",
        "Video": "mp4 avi mkv webm gif mov",
        "Image": "jpg jpeg png webp",
}.items():
    prefix = f"Test{kind}Converter_Convert_Integration_Exhaustive/"
    expected = {f"{prefix}.{src}->.{dst}_({quality})"
                for src in formats.split() for dst in formats.split()
                for quality in ("High", "Balanced", "Low")}
    actual = {name for name in passed if name.startswith(prefix)}
    if actual != expected:
        raise SystemExit(f"{kind}: missing {sorted(expected - actual)}, unexpected {sorted(actual - expected)}")
    print(f"| {kind} | {len(expected)} | {len(actual)} |")

covered = total = 0
for line in profile.read_text().splitlines()[1:]:
    _, statements, hits = line.rsplit(" ", 2)
    total += int(statements)
    if int(hits):
        covered += int(statements)
print(f"\nGo statement coverage: **{covered / total:.1%}** ({covered}/{total}).")
