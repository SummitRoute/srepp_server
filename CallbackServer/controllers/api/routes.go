////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package api

import (
	"net/http"

	_ "github.com/lib/pq" // Needed for gorp
	"github.com/zenazn/goji/web"

	"qdserver/CallbackServer/system"
)

// Controller object
type Controller struct {
	system.Controller
}

// Status health check for the load balancer
func (controller *Controller) Status(c web.C, r *http.Request) (string, int) {
	return "Summit Route callback server", http.StatusOK
}
