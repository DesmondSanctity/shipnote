# Demo casts

Short asciinema recordings used in the README and on socials when announcing
releases. **We use [asciinema](https://asciinema.org) — not VHS.**

Recording a new cast (≤ 25 rows, ≤ 100 cols keeps it readable on Twitter/X
and embedded in the README):

```sh
asciinema rec --cols 100 --rows 25 docs/casts/init.cast
asciinema rec --cols 100 --rows 25 docs/casts/generate.cast
asciinema rec --cols 100 --rows 25 docs/casts/site.cast
```

Convert a cast to a GIF for embedding (uses [agg](https://github.com/asciinema/agg)):

```sh
make casts
```

That writes `docs/img/<name>.gif` for every `.cast` file in this directory.

Posting on socials? Re-upload the `.cast` to https://asciinema.org for an
interactive embed, and link the `.gif` for platforms that don't render
asciinema (X, LinkedIn).
