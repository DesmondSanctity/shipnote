package cli

import (
	"io"
	"os"
)

// banner is the ASCII art shown on bare `shipnote` invocation.
// Keep it short so it doesn't dominate the terminal.
const banner = `
       _     _                   _
   ___| |__ (_)_ __  _ __   ___ | |_ ___
  / __| '_ \| | '_ \| '_ \ / _ \| __/ _ \
  \__ \ | | | | |_) | | | | (_) | ||  __/
  |___/_| |_|_| .__/|_| |_|\___/ \__\___|
              |_|  release intelligence
`

// noColor reports whether banner/colorized output should be suppressed.
// Honors the NO_COLOR convention (https://no-color.org) and CI=true.
func noColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	return os.Getenv("CI") == "true"
}

// printBanner writes the banner to w. Currently the art is plain ASCII;
// noColor() exists so we can later add color codes without breaking CI.
func printBanner(w io.Writer) {
	_, _ = io.WriteString(w, banner)
}
