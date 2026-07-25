# Notes

Scratch documentation that has nothing to do with `services/api` or
`lib/common`. It exists to demonstrate the KP memo: edit this file, rebuild,
and the build step reports **(cached)** — the file is outside the covered
closure, so the build's identity (KP) is unchanged.

Try it:

```bash
echo "- edited $(date -u +%FT%TZ)" >> docs/notes.md
jobs-client build --source services/api     # ✓ build …  (cached)
```
