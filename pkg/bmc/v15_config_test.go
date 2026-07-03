package bmc

import "testing"

func TestParseV15AuthTypes(t *testing.T) {
	types, err := ParseV15AuthTypes("md5, MD2, password")
	if err != nil {
		t.Fatalf("ParseV15AuthTypes: %v", err)
	}
	if len(types) != 3 {
		t.Fatalf("want 3 types, got %v", types)
	}
	if types[0] != V15AuthTypeMD5 || types[1] != V15AuthTypeMD2 || types[2] != V15AuthTypePassword {
		t.Fatalf("unexpected types: %v", types)
	}
}

func TestFormatV15AuthTypes(t *testing.T) {
	got := FormatV15AuthTypes([]V15AuthType{V15AuthTypeMD5, V15AuthTypeMD2})
	if got != "md5,md2" {
		t.Fatalf("FormatV15AuthTypes: got %q", got)
	}
}
