package comfyui

type ComfyQueuePromptRes struct {
	PromptID   string   `json:"prompt_id"`
	Number     int      `json:"number"`
	NodeErrors struct{} `json:"node_errors"`
}

type NodeData struct {
	ClassType string `json:"class_type,omitempty"`
}

type ComfyImageUploadRes struct {
	Name      string `json:"name"`
	SubFolder string `json:"subfolder"`
	Type      string `json:"type"`
}

// 结构体定义
type ImageInfo struct {
	Filename  string `json:"filename"`
	SubFolder string `json:"subfolder"`
	Type      string `json:"type"`
}

type Output struct {
	Images []*ImageInfo `json:"images"`
	Text   []string     `json:"text"`
}

type HistoryData struct {
	Outputs map[string]*Output `json:"outputs"`
}

type QueueRes struct {
	QueueRunning []interface{} `json:"queue_running"`
	QueuePending []interface{} `json:"queue_pending"`
}
