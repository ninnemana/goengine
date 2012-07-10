package controllers

var (
	controllerlist = make(map[string]*Controller)
)

type Action struct {
	name string
	run  func() map[string]interface{}
}

type Controller struct {
	name    string
	actions map[string]*Action
}

func init() {

	controllerlist["home"] = &Controller{
		name:    "home",
		actions: homeActions(),
	}
}
