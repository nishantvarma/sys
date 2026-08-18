type tangler struct {
	chunks   map[string]chunk
	placed   map[string]bool
	swept    map[string]bool
	expanded map[string]bool
}
