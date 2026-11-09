package list

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
	want := "ID" + strings.Repeat(" ", 22) + " : 1\n" +
		"Name" + strings.Repeat(" ", 20) + " : Fan\n" +
		"\n" +
		"ID" + strings.Repeat(" ", 22) + " : 2\n" +
		"Name" + strings.Repeat(" ", 20) + " : PSU\n"
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
	want := "ID" + strings.Repeat(" ", 22) + " : 1\n" +
		"SensorName" + strings.Repeat(" ", 14) + " : CPU1 Temp\n"
	if b.String() != want {
		t.Fatalf("got %q, want %q", b.String(), want)
	}
}
