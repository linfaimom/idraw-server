package response

type RecordDto struct {
	Id     uint     `json:"id"`
	Type   string   `json:"type"`
	Input  string   `json:"input"`
	Output []string `json:"output"`
}
