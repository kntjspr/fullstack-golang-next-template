package swagger

import "testing"

func TestExtractOpenAPITitle(t *testing.T) {
	tests := []struct {
		name string
		spec string
		want string
	}{
		{
			name: "quoted title with colon",
			spec: "openapi: 3.0.3\ninfo:\n  title: \"My API: v2\"\n",
			want: "My API: v2",
		},
		{
			name: "missing title fallback",
			spec: "openapi: 3.0.3\ninfo:\n  version: 1.0.0\n",
			want: "API Docs",
		},
		{
			name: "invalid yaml fallback",
			spec: "openapi: [invalid",
			want: "API Docs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractOpenAPITitle([]byte(tc.spec))
			if got != tc.want {
				t.Fatalf("unexpected title: got %q want %q", got, tc.want)
			}
		})
	}
}
