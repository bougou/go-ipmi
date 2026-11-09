package json

import (
	"slices"
	"strings"
	"testing"

	"github.com/bougou/go-ipmi/pkg/render"
)

func TestRender(t *testing.T) {
	r := Renderer{}
	got := r.Render(
		render.Columns("ID", "Name"),
		[]map[string]string{{"ID": "1", "Name": "Fan"}, {"ID": "2", "Name": "PSU"}},
	)
	want := `[
  {
    "ID": "1",
    "Name": "Fan"
  },
  {
    "ID": "2",
    "Name": "PSU"
  }
]` + "\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderEmpty(t *testing.T) {
	r := Renderer{}
	if got := r.Render(render.Columns("ID", "Name"), nil); got != "[]\n" {
		t.Fatalf("got %q, want %q", got, "[]\n")
	}
}

func TestRenderEscapes(t *testing.T) {
	r := Renderer{}
	got := r.Render(render.Columns("A"), []map[string]string{{"A": "x\"y\nz"}})
	want := `[
  {
    "A": "x\"y\nz"
  }
]` + "\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderStreamIgnoresWidth(t *testing.T) {
	var b strings.Builder
	rows := slices.Values([]map[string]string{{"ID": "1", "SensorName": "CPU1 Temp"}})
	columns := []render.Column{{Name: "ID"}, {Name: "SensorName", Width: 40}}
	if err := (Renderer{}).RenderStream(&b, columns, rows); err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "ID": "1",
    "SensorName": "CPU1 Temp"
  }
]` + "\n"
	if b.String() != want {
		t.Fatalf("got %q, want %q", b.String(), want)
	}
}
