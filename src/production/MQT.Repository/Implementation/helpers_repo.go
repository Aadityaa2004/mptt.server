package implementation

// ensureMetaNotNull ensures Meta is not nil to prevent null JSON issues
func ensureMetaNotNull(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return make(map[string]interface{})
	}
	return meta
}