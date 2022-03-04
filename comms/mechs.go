package comms

import (
	"github.com/ninja-software/terror/v2"
)

func (c *C) GetMechOwner() error {

	return terror.ErrNotImplemented
}

func (c *C) MigrateOnlyGetAllMechs(req MigrateOnlyGetAllMechsReq, resp *MigrateOnlyGetAllMechsResp) error {
	return terror.ErrNotImplemented
}

func (c *C) MigrateOnlyGetAllMechTemplates(req MigrateOnlyGetAllTemplatesReq, resp *MigrateOnlyGetAllTemplatesResp) error {
	return terror.ErrNotImplemented
}
