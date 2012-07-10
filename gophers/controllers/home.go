package controllers

func homeActions() map[string]*Action {

	actionlist := make(map[string]*Action)

	actionlist["index"] = &Action{
		Name: "index",
		Run:  homeIndex(),
	}

	return actionlist
}

func homeIndex() map[string]interface{} {
	vb := make(map[string]interface{})
	return vb
}
