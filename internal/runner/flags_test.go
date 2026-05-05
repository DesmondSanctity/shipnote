package runner_test

import "flag"

// update lets `go test -update` rewrite the checked-in golden files
// rather than asserting against them. Off by default.
var update = flag.Bool("update", false, "update golden files")
