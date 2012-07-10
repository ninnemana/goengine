package controllers

func loadActionByName(controller string, action string) {
	eval(controller + "." + action + "()")
}
