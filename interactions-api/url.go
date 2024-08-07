package interactionsapi

import (
	"net/url"
)

func sameOrigin(a, b url.URL) bool {
	return a.Scheme == b.Scheme &&
		a.User == b.User &&
		a.Host == b.Host
}
