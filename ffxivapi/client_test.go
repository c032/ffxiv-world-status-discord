package ffxivapi_test

import (
	"testing"

	"github.com/c032/ffxiv-world-status-discord/ffxivapi"
)

var options = ffxivapi.ClientOptions{
	BaseURL: "https://ffxiv.c032.dev/api/",
}

func TestClient_Worlds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	c, err := ffxivapi.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}

	worldsResponse, err := c.Worlds()
	if err != nil {
		t.Fatal(err)
	}

	if worldsResponse == nil {
		t.Fatalf("c.Worlds() = nil; want non-nil")
	}

	if len(worldsResponse.Worlds) == 0 {
		t.Fatalf("len(c.Worlds().Worlds) = 0; want at least one")
	}
}
