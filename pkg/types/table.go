package types

import (
	"iter"
	"os"

	"github.com/bougou/go-ipmi/pkg/render"
	"github.com/bougou/go-ipmi/pkg/render/ascii"
)

// defaultRenderer is used by the Format* helpers when no renderer is given.
var defaultRenderer = ascii.Renderer{}

// RenderTable formats rows with the default (ASCII-bordered) renderer.
func RenderTable(headers []string, rows []map[string]string) string {
	return ascii.Renderer{}.Render(render.Columns(headers...), rows)
}

// RenderTableStream streams rows to stdout with the default (ASCII-bordered) renderer.
func RenderTableStream(headers []string, rows iter.Seq[map[string]string]) error {
	return ascii.Renderer{}.RenderStream(os.Stdout, render.Columns(headers...), rows)
}
