package gophers

var (
	controllerlist = make(map[string]*Controller)
)

type Action struct {
	Name     string
	Run      map[string]interface{}
	Template string
	Layout   string
}

type Controller struct {
	Name    string
	Actions map[string]*Action
}

func GenerateControllers(ctx WebContext) map[string]*Controller {

	controllerlist["home"] = &Controller{
		Name:    "home",
		Actions: homeActions(ctx),
	}

	return controllerlist
}
