package features

import (
	"net/http"

	"github.com/oullin/alloy/pkg/hub/inertia/protocol"
)

func (a app) linkPrefetchHandler(w http.ResponseWriter, r *http.Request) {
	a.container.Render(w, r, "Features/Prefetching/LinkPrefetch", protocol.Props{})
}

func (a app) staleWhileRevalidateHandler(w http.ResponseWriter, r *http.Request) {
	a.container.Render(w, r, "Features/Prefetching/StaleWhileRevalidate", protocol.Props{})
}

func (a app) manualPrefetchHandler(w http.ResponseWriter, r *http.Request) {
	a.container.Render(w, r, "Features/Prefetching/ManualPrefetch", protocol.Props{})
}

func (a app) cacheManagementHandler(w http.ResponseWriter, r *http.Request) {
	a.container.Render(w, r, "Features/Prefetching/CacheManagement", protocol.Props{})
}
