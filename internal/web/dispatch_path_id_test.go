package web

import (
	"net/http/httptest"
	"testing"
)

func TestDispatchPathID(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		prefix  string
		wantID  string
		wantSet bool
	}{
		{
			name:    "recipient edit",
			path:    "/recipients/abc-123",
			prefix:  "/recipients/",
			wantID:  "abc-123",
			wantSet: true,
		},
		{
			name:    "recipient secrets management",
			path:    "/recipients/23e36417-16b4-483b-8682-a4528c09d636/secrets",
			prefix:  "/recipients/",
			wantID:  "23e36417-16b4-483b-8682-a4528c09d636",
			wantSet: true,
		},
		{
			// The /test action was removed (the only thing it did was
			// re-send the hello email; that's now create-time-only via
			// the send_intro checkbox). The URL no longer routes
			// anywhere meaningful, but path extraction must keep
			// returning the id segment so any unintended hit falls
			// through to the edit handler safely.
			name:    "trailing segment ignored — id is still first segment",
			path:    "/recipients/abc-123/whatever",
			prefix:  "/recipients/",
			wantID:  "abc-123",
			wantSet: true,
		},
		{
			name:    "secret view",
			path:    "/secrets/sec-1",
			prefix:  "/secrets/",
			wantID:  "sec-1",
			wantSet: true,
		},
		{
			name: "secret assign — id is the first segment, not the last " +
				"(GetLastURLSegment would have returned 'assign')",
			path:    "/secrets/sec-1/assign",
			prefix:  "/secrets/",
			wantID:  "sec-1",
			wantSet: true,
		},
		{
			name:    "missing id — bare prefix",
			path:    "/recipients/",
			prefix:  "/recipients/",
			wantID:  "",
			wantSet: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			got := dispatchPathID(req, tc.prefix)
			if got != tc.wantID {
				t.Fatalf("dispatchPathID returned %q, want %q", got, tc.wantID)
			}
			if got != req.PathValue("id") {
				t.Fatalf("PathValue(\"id\") = %q, want %q (function returned %q)",
					req.PathValue("id"), tc.wantID, got)
			}
			if tc.wantSet && req.PathValue("id") == "" {
				t.Fatal("PathValue(\"id\") not set; downstream handlers will respond 400 \"ID is required\"")
			}
		})
	}
}
