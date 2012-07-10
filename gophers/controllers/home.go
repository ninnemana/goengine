package controllers

func homeActions() map[string]*Action {

	actionlist := make(map[string]*Action)

	actionlist["index"] = &Action{
		name: "index",
		run: func() map[string]interface{} {
			vb := make(map[string]interface{})
			return vb
		},
	}

	return actionlist
}
