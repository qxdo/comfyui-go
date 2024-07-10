package comfy_tasks

import (
	"context"
	"github.com/qxdo/comfyui-go/comfyui"
)

const (
	ExecutionError  = "execution_error"
	ExecutionStart  = "execution_start"
	Executed        = "executed"
	Executing       = "executing"
	Status          = "status"
	ExecutionCached = "execution_cached"
)

var SupportTypeMap = map[string]struct{}{
	ExecutionError:  {},
	ExecutionStart:  {},
	Executed:        {},
	Executing:       {},
	Status:          {},
	ExecutionCached: {},
}

// http endpoint like : https://127.0.0.1:8443
// ws endpoint like: ws://127.0.0.1:8000/ws?clientId=1234
// 这里是comfyUI执行任务的抽象，每个任务只需要定义并且实现这几个方法 即可将comfyUI的流程通用化处理
type AigcTask interface {
	// 任务ID 拿到任务的id
	GetTaskID(ctx context.Context) string
	// 取到 extra_data的方法
	GetExtraData(ctx context.Context) string
	// prompt params处理方法
	GetPrompt(ctx context.Context) map[string]interface{}
	// 任务失败处理方法
	TaskFailed(ctx context.Context, reason, errMsg string) error
	// queue prompt 操作后置方法
	AfterQueuePrompt(ctx context.Context, promptID string) error
	// 在进入ws之前 check一下任务状态
	BeforeWebSocketCheck(ctx context.Context, taskID string) (int, error)
	// 解析WebSocket传进来的bin数据
	ParseBinData(ctx context.Context, message []byte) error
	// comfy 解析 hook
	ExecutionError(ctx context.Context, data *ServData) (bool, error)
	// comfy 解析 hook
	ExecutionStart(ctx context.Context, data *ServData) (bool, error)
	// comfy 解析 hook
	ExecutedImages(ctx context.Context, data *ServData) (bool, error)
	// comfy 解析 hook
	ExecutedText(ctx context.Context, data *ServData) (bool, error)
	// comfy 解析 hook
	Executing(ctx context.Context, data *ServData) (bool, error)
	// 拿到ticker等待的时间
	GetTaskTimeoutTickerTime() int
	// 记录发给comfy的json数据
	RecordParams(ctx context.Context, jsonParams string) error
}

type MessageData struct {
	Type string `json:"type,omitempty"`
}

type ServInnerData struct {
	Node        string      `json:"node,omitempty"`
	PromptID    string      `json:"prompt_id,omitempty"`
	ExecMessage string      `json:"exception_message,omitempty"`
	Output      *OutputData `json:"output,omitempty"`
}

type OutputData struct {
	Images []*comfyui.ImageInfo `json:"images,omitempty"`
	Text   []string             `json:"text,omitempty"`
	Tags   []string             `json:"tags,omitempty"`
}

// {"node": "14", "output": {"images": [{"filename": "ComfyUI_temp_fhjpt_00009_.png", "subfolder": "", "type": "temp"}, {"filename": "ComfyUI_temp_fhjpt_00010_.png", "subfolder": "", "type": "temp"}, {"filename": "ComfyUI_temp_fhjpt_00011_.png", "subfolder": "", "type": "temp"}, {"filename": "ComfyUI_temp_fhjpt_00012_.png", "subfolder": "", "type": "temp"}]}
type ServData struct {
	Type string         `json:"type,omitempty"`
	Data *ServInnerData `json:"data,omitempty"`
	// 以下为自己带的参数
	TaskID         string              `json:"task_id,omitempty"`
	HttpEndPoint   string              `json:"http_end_point,omitempty"`
	ProcessData    []string            `json:"process_data,omitempty"`
	PreviewNodeMap map[string]struct{} `json:"preview_node_map,omitempty"`
}
