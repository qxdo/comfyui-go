package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/qxdo/comfyui-go/comfy_tasks"
	"github.com/qxdo/comfyui-go/comfyui"
	"os"
)

type MyAigcTask struct {
	TaskID string `json:"task_id"`
	Params string `json:"params"`
}

func New(taskID, params string) *MyAigcTask {
	return &MyAigcTask{
		TaskID: taskID,
		Params: params,
	}
}

func (m *MyAigcTask) RecordParams(ctx context.Context, jsonParams string) error {
	_ = os.WriteFile(uuid.NewString()+"_comfy_queue_prompt_params.json", []byte(jsonParams), os.ModePerm)
	return nil
}

func (m *MyAigcTask) GetTaskID(ctx context.Context) string {
	return m.TaskID
}

func (m *MyAigcTask) GetPrompt(ctx context.Context) map[string]interface{} {
	var data interface{}
	_ = json.Unmarshal([]byte(m.Params), &data)
	extraData := m.GetExtraData(ctx)
	var dataMap = map[string]interface{}{
		"prompt": data,
	}

	if extraData != "" {
		var data1 interface{}
		_ = json.Unmarshal([]byte(extraData), &data1)
		dataMap["extra_data"] = data1
	}
	return dataMap
}

func (m *MyAigcTask) TaskFailed(ctx context.Context, reason, errMsg string) error {
	fmt.Println("task failed:", m.TaskID, reason, errMsg)
	return nil
}

// queue prompt 操作后置方法
func (m *MyAigcTask) AfterQueuePrompt(ctx context.Context, promptID string) error {
	fmt.Println("after queue prompt:", promptID)
	return nil
}

// 在进入ws之前 check一下任务状态 要至少返回1 才能继续走下去
func (m *MyAigcTask) BeforeWebSocketCheck(ctx context.Context, taskID string) (int, error) {
	fmt.Println("BeforeWebSocketCheck taskID:", taskID)
	return 1, nil
}

// 解析WebSocket传进来的bin 数据 目前是没有用的
func (m *MyAigcTask) ParseBinData(ctx context.Context, message []byte) error {
	fmt.Println("ParseBinData message  len:", len(message))
	return nil
}

func (m *MyAigcTask) GetExtraData(ctx context.Context) string {
	return ""
}

// comfy 解析 hook
func (m *MyAigcTask) ExecutionError(ctx context.Context, data *comfy_tasks.ServData) (bool, error) {
	fmt.Println("ExecutionError data:", data)
	return true, nil
}

// comfy 解析 hook
func (m *MyAigcTask) ExecutionStart(ctx context.Context, data *comfy_tasks.ServData) (bool, error) {
	fmt.Println("ExecutionStart data:", data)
	return false, nil
}

func (m *MyAigcTask) ExecutedImages(ctx context.Context, data *comfy_tasks.ServData) (bool, error) {
	fmt.Println("Executed Image data:", data, "image len:", len(data.Data.Output.Images))
	for _, image := range data.Data.Output.Images {
		fmt.Println(comfyui.GetComfyPreviewLink(ctx, data.HttpEndPoint, image))
	}
	return true, nil
}

// 生成文本内容
func (m *MyAigcTask) ExecutedText(ctx context.Context, data *comfy_tasks.ServData) (bool, error) {
	fmt.Println("Executed Text data:", data)
	fmt.Println("exec len:", len(data.Data.Output.Text))
	for _, s := range data.Data.Output.Text {
		fmt.Println(s)
	}
	return true, nil
}

// comfy 解析 hook
func (m *MyAigcTask) Executing(ctx context.Context, data *comfy_tasks.ServData) (bool, error) {
	fmt.Println("Executing data:", data)
	return false, nil
}

// 拿到ticker等待的时间 比如 第一个项目像等待2分钟 第二个想等待4分钟 可以自定义配置 单位:秒
func (m *MyAigcTask) GetTaskTimeoutTickerTime() int {
	return 240
}
