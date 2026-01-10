package ibmmq

import (
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/ibmmq", new(RootModule))
}

type RootModule struct{}

type ModuleInstance struct {
	vu modules.VU
}

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{vu: vu}
}

func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"QueueManager": mi.NewQueueManager,
		},
	}
}
