package gophers

func homeActions(ctx WebContext) map[string]*Action {

	actionlist := make(map[string]*Action)

	actionlist["index"] = &Action{
		Name: "index",
		Run:  homeIndex(ctx),
	}

	return actionlist
}

func homeIndex(ctx WebContext) map[string]interface{} {
	vb := make(map[string]interface{})
	return vb
}
