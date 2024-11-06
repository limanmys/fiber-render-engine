package helpers

// This function x string maps together
func MergeStringMaps(ms ...map[string]string) map[string]string {
	res := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// This function x interface maps together
func MergeInterfaceMaps(ms ...map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}
