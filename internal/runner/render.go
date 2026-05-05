package runner

import (
	"github.com/DesmondSanctity/shipnote/internal/model"
	"github.com/DesmondSanctity/shipnote/internal/render"
)

// Indirection so the heavy renderer lives in its own file and the
// orchestrator stays focused. Both are pure.
func renderMarkdown(r model.Release) string { return render.Markdown(r) }

func renderJSON(r model.Release) ([]byte, error) { return render.JSON(r) }
