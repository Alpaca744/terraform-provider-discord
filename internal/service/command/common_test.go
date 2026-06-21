package command

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func TestOptionsSemanticEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b jsontypes.Normalized
		want bool
	}{
		{
			name: "both null",
			a:    jsontypes.NewNormalizedNull(),
			b:    jsontypes.NewNormalizedNull(),
			want: true,
		},
		{
			name: "null vs empty string",
			a:    jsontypes.NewNormalizedNull(),
			b:    jsontypes.NewNormalizedValue(""),
			want: true,
		},
		{
			name: "configured vs api echo with extra null keys",
			a:    jsontypes.NewNormalizedValue(`[{"type":1,"name":"set","description":"set"}]`),
			b:    jsontypes.NewNormalizedValue(`[{"name":"set","description":"set","type":1,"name_localizations":null,"description_localizations":null}]`),
			want: true,
		},
		{
			name: "key order and whitespace differences",
			a:    jsontypes.NewNormalizedValue(`[{"name":"a","type":3}]`),
			b:    jsontypes.NewNormalizedValue(`[ { "type": 3, "name": "a" } ]`),
			want: true,
		},
		{
			name: "nested null keys stripped recursively",
			a:    jsontypes.NewNormalizedValue(`[{"type":2,"name":"grp","options":[{"type":1,"name":"sub"}]}]`),
			b:    jsontypes.NewNormalizedValue(`[{"type":2,"name":"grp","options":[{"type":1,"name":"sub","name_localizations":null}]}]`),
			want: true,
		},
		{
			name: "genuine value difference",
			a:    jsontypes.NewNormalizedValue(`[{"type":1,"name":"set","description":"set"}]`),
			b:    jsontypes.NewNormalizedValue(`[{"type":1,"name":"set","description":"changed"}]`),
			want: false,
		},
		{
			name: "configured options vs none",
			a:    jsontypes.NewNormalizedValue(`[{"type":1,"name":"set"}]`),
			b:    jsontypes.NewNormalizedNull(),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := optionsSemanticEqual(tc.a, tc.b); got != tc.want {
				t.Errorf("optionsSemanticEqual(%q, %q) = %v, want %v",
					tc.a.ValueString(), tc.b.ValueString(), got, tc.want)
			}
		})
	}
}
