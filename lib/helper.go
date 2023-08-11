package lib

func TMI(item interface{}) map[string]interface{} {
	return item.(map[string]interface{})
}
