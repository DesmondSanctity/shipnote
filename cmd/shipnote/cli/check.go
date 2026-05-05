package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// checkArtifacts compares freshly rendered output against on-disk
// files. Exits with an error (translated to exit code 1 by main) if
// either file is missing or differs.
func checkArtifacts(cmd *cobra.Command, mdPath, jsPath, gotMD string, gotJSON []byte) error {
	if drift, err := fileDiffers(mdPath, []byte(gotMD)); err != nil {
		return err
	} else if drift {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "shipnote: %s is out of date\n", mdPath)
		return errCheckDrift
	}
	if drift, err := fileDiffers(jsPath, gotJSON); err != nil {
		return err
	} else if drift {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "shipnote: %s is out of date\n", jsPath)
		return errCheckDrift
	}
	return nil
}

func fileDiffers(path string, want []byte) (bool, error) {
	got, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	return !bytesEqual(got, want), nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// errCheckDrift signals the --check flag detected a regen-diff. Kept
// as a sentinel so main can map it to exit code 1 distinctly from
// other errors if needed later.
var errCheckDrift = fmt.Errorf("regenerated output differs from on-disk artifacts")
