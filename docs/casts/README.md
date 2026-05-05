# Demo casts

Asciicast v2 files used to render the GIFs in the README and on socials.

We **synthesize** the casts deterministically from [build.py](build.py)
rather than recording a live session. That keeps the timing reproducible
and lets us tweak copy without re-recording. agg(1) renders them to GIF
the same way it would real recordings.

## Workflow

```sh
# 1. Edit casts in build.py (banner, prompts, output, timing)
$EDITOR docs/casts/build.py

# 2. Regenerate .cast files
python3 docs/casts/build.py

# 3. Render GIFs into docs/img/
make casts        # requires `agg` — brew install agg
```

## Files

| Cast            | GIF                     | Purpose                                                          |
| --------------- | ----------------------- | ---------------------------------------------------------------- |
| `demo.cast`     | `docs/img/demo.gif`     | Hero GIF for the README + social announcements.                  |
| `init.cast`     | `docs/img/init.gif`     | `shipnote init` flow.                                            |
| `generate.cast` | `docs/img/generate.gif` | `shipnote generate --ai` + a peek at the resulting CHANGELOG.md. |
| `site.cast`     | `docs/img/site.gif`     | `shipnote site` flow + output layout.                            |

## Posting on socials

Re-upload the `.cast` to https://asciinema.org for an interactive embed,
and link the `.gif` for platforms that don't render asciinema (X,
LinkedIn).

## Style

Mirrors the loadam demo style: orange (256-color 208) hero banner, dim
tagline, `$` prompt, ~40 ms/char typing, 100×30 viewport. Adjust in
`build.py`.
