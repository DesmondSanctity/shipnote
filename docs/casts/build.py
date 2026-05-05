#!/usr/bin/env python3
"""Generate asciicast v2 .cast files for the shipnote demo GIFs.

Style mirrors github.com/desmondsanctity/loadam: orange hero banner,
dim tagline, $ prompt, ~40ms/char typing. Casts are hand-rolled
(JSONL: header + [time, "o", text] events) so output is deterministic
and reproducible without a real recording session.

Run:
    python3 docs/casts/build.py
    make casts          # renders docs/img/<name>.gif via agg
"""

import json
import pathlib

ROOT = pathlib.Path(__file__).resolve().parent
COLS, ROWS = 100, 30

ORANGE = "\x1b[38;5;208m"
GREEN = "\x1b[32m"
CYAN = "\x1b[36m"
DIM = "\x1b[2m"
BOLD = "\x1b[1m"
RED = "\x1b[31m"
YELLOW = "\x1b[33m"
RESET = "\x1b[0m"

PROMPT = f"{GREEN}${RESET} "

BANNER = (
    f"{ORANGE}"
    "       _     _                   _\r\n"
    "   ___| |__ (_)_ __  _ __   ___ | |_ ___\r\n"
    "  / __| '_ \\| | '_ \\| '_ \\ / _ \\| __/ _ \\\r\n"
    "  \\__ \\ | | | | |_) | | | | (_) | ||  __/\r\n"
    "  |___/_| |_|_| .__/|_| |_|\\___/ \\__\\___|\r\n"
    "              |_|\r\n"
    f"{RESET}{DIM}  commits → categorized release notes · markdown · ai · public site{RESET}\r\n"
)


def write_cast(name, frames, title):
    path = ROOT / f"{name}.cast"
    with path.open("w", encoding="utf-8") as f:
        header = {
            "version": 2,
            "width": COLS,
            "height": ROWS,
            "timestamp": 1777632900,
            "env": {"SHELL": "/bin/zsh", "TERM": "xterm-256color"},
            "title": title,
        }
        f.write(json.dumps(header) + "\n")
        t = 0.0
        for delay, text in frames:
            t = round(t + delay, 3)
            f.write(json.dumps([t, "o", text]) + "\n")
    print(f"wrote docs/casts/{name}.cast")


def banner_frames():
    return [(0, BANNER), (0.6, "")]


def type_cmd(cmd, per_char=0.04, pause_before=0.3, pause_after=0.4):
    out = [(pause_before, PROMPT)]
    for i, ch in enumerate(cmd):
        out.append((per_char if i else 0.0, ch))
    out.append((pause_after, "\r\n"))
    return out


def line(delay, text=""):
    return (delay, text + "\r\n")


def demo_hero():
    """Single hero cast for the README — banner + headline commands."""
    frames = banner_frames()
    frames += type_cmd("shipnote generate --ai")
    frames += [
        line(0.25, f"{DIM}shipnote: collecting v0.1.0..HEAD{RESET}"),
        line(0.18, f"{DIM}shipnote: 14 PRs, 31 commits, 4 contributors{RESET}"),
        line(0.18, f"{DIM}shipnote: categorized  feat=5 fix=6 perf=1 breaking=2{RESET}"),
        line(0.30, f"{DIM}shipnote: ai (anthropic/claude-3-5-haiku) cache 9/14, generated 5{RESET}"),
        line(0.20, f"{CYAN}\u2713{RESET} CHANGELOG.md  {DIM}+128 \u22120{RESET}"),
        line(0.10, f"{CYAN}\u2713{RESET} .shipnote/release.json  {DIM}schema v1.0.0{RESET}"),
        (0.5, ""),
    ]
    frames += type_cmd("shipnote site public/changelog --site-url https://acme.com/changelog")
    frames += [
        line(0.25, f"{DIM}shipnote: rendering 8 releases \u00b7 feed.json \u00b7 feed.xml{RESET}"),
        line(0.20, f"{CYAN}\u2713{RESET} wrote site to {BOLD}public/changelog{RESET}"),
        line(0.05, f"  {DIM}drop on GitHub Pages, S3, Netlify \u2014 it's static.{RESET}"),
        (0.4, PROMPT),
        (1.8, ""),
    ]
    write_cast("demo", frames, "shipnote in 30 seconds")


def demo_init():
    frames = banner_frames()
    frames += type_cmd("shipnote init --yes")
    frames += [
        line(0.20, f"{DIM}shipnote v0.1.0{RESET}"),
        line(0.12, f"{CYAN}\u2713{RESET} detected GitHub remote: acme/widgets"),
        line(0.10, f"{CYAN}\u2713{RESET} wrote .shipnote.toml"),
        line(0.10, f"{CYAN}\u2713{RESET} added .shipnote/ to .gitignore"),
        line(0.10, ""),
        line(0.05, f"  ready. run {BOLD}shipnote generate{RESET} to render your first changelog."),
        (0.4, PROMPT),
        (1.8, ""),
    ]
    write_cast("init", frames, "shipnote init")


def demo_generate():
    frames = banner_frames()
    frames += type_cmd("shipnote generate --ai")
    frames += [
        line(0.25, f"{DIM}shipnote: collecting v0.1.0..HEAD{RESET}"),
        line(0.18, f"{DIM}shipnote: 14 PRs, 31 commits, 4 contributors{RESET}"),
        line(0.18, f"{DIM}shipnote: categorized  feat=5 fix=6 perf=1 breaking=2{RESET}"),
        line(0.30, f"{DIM}shipnote: ai (anthropic/claude-3-5-haiku) cache 9/14, generated 5{RESET}"),
        line(0.20, f"{CYAN}\u2713{RESET} wrote {BOLD}CHANGELOG.md{RESET} and {BOLD}.shipnote/release.json{RESET}"),
        (0.4, ""),
    ]
    frames += type_cmd("head -10 CHANGELOG.md")
    frames += [
        line(0.12, f"{BOLD}# v0.2.0 \u2014 2026-05-05{RESET}"),
        line(0.04, ""),
        line(0.04, f"{YELLOW}Adds team workspaces and tightens breaking-change checks.{RESET}"),
        line(0.04, ""),
        line(0.04, f"{RED}{BOLD}## Breaking Changes{RESET}"),
        line(0.04, f"- {BOLD}auth:{RESET} drop legacy session cookie {DIM}#142{RESET}"),
        line(0.04, f"- {BOLD}api:{RESET} require ?team_id on /v1/events {DIM}#151{RESET}"),
        line(0.04, ""),
        line(0.04, f"{GREEN}{BOLD}## Added{RESET}"),
        line(0.04, f"- team workspaces with role-based access {DIM}#138{RESET}"),
        line(0.04, f"- audit log export to CSV {DIM}#147{RESET}"),
        (0.4, PROMPT),
        (1.8, ""),
    ]
    write_cast("generate", frames, "shipnote generate")


def demo_site():
    frames = banner_frames()
    frames += type_cmd("shipnote site public/changelog --site-url https://acme.com/changelog")
    frames += [
        line(0.20, f"{DIM}shipnote: collecting v0.1.0..HEAD{RESET}"),
        line(0.15, f"{DIM}shipnote: rendering 8 releases{RESET}"),
        line(0.15, f"{DIM}shipnote: writing JSON Feed 1.1 + Atom{RESET}"),
        line(0.10, f"{CYAN}\u2713{RESET} wrote site to {BOLD}public/changelog{RESET}"),
        (0.4, ""),
    ]
    frames += type_cmd("ls public/changelog")
    frames += [
        line(0.05, f"{CYAN}assets{RESET}        feed.xml      releases.json"),
        line(0.05, f"feed.json     index.html    {CYAN}v0.1.0{RESET}        {CYAN}v0.2.0{RESET}"),
        line(0.10, ""),
        line(0.05, f"  {DIM}embed widget on any page \u2192 done.{RESET}"),
        (0.4, PROMPT),
        (1.8, ""),
    ]
    write_cast("site", frames, "shipnote site")


if __name__ == "__main__":
    demo_hero()
    demo_init()
    demo_generate()
    demo_site()
    print("\nnext: make casts   # renders docs/img/<name>.gif via agg")
