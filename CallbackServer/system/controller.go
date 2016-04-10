////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////

package system

import (
	"github.com/coopernurse/gorp"
	"github.com/streadway/amqp"
	"github.com/zenazn/goji/web"
)

// Controller struct
type Controller struct {
}

// GetDatabase helper
func (controller *Controller) GetDatabase(c web.C) *gorp.DbMap {
	if db, ok := c.Env["DBSession"].(*gorp.DbMap); ok {
		return db
	}

	return nil
}

// GetQueueChannel helper
func (controller *Controller) GetQueueChannel(c web.C) *amqp.Channel {
	if ch, ok := c.Env["QueueChannel"].(*amqp.Channel); ok {
		return ch
	}

	return nil
}
