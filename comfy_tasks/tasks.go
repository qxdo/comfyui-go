package comfy_tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/qxdo/comfyui-go/comfyui"
	"os"
	"os/signal"
	"time"
)

type aigc struct {
	ws           *websocket.Conn
	errChan      chan error
	clientID     string
	wsEndPoint   string
	httpEndpoint string
	ctx          context.Context
	task         AigcTask
	logger       Logger
}

func New(ctx context.Context, httpEndpoint, wsEndpoint string, task AigcTask, logger Logger) *aigc {
	clientID := uuid.NewString()
	return &aigc{
		errChan:      make(chan error),
		clientID:     clientID,
		httpEndpoint: httpEndpoint,
		wsEndPoint:   fmt.Sprintf(comfyui.ComfyWSLink, wsEndpoint, clientID),
		ws:           nil,
		ctx:          ctx,
		task:         task,
		logger:       logger,
	}
}

func (a *aigc) close() {
	if a.ws != nil {
		_ = a.ws.Close()
	}
	a.logger.Info(a.ctx, "aigc server closed")
}

func (a *aigc) Start() error {
	// 主流程
	err := a.connectWebSocketConn()
	if err != nil {
		err = a.task.TaskFailed(a.ctx, "start_web_sock_error", err.Error())
		if err != nil {
			a.logger.Info(a.ctx, "task failed")
			return err
		}
		return err
	}
	var promptID string
	promptID, err = a.queuePrompt()
	if err != nil {
		a.logger.Info(a.ctx, "queuePrompt error:", err)
		a.errChan <- err
		return err
	}
	err = a.task.AfterQueuePrompt(a.ctx, promptID)
	if err != nil {
		a.logger.Info(a.ctx, "AfterQueuePrompt error:", err)
		return err
	}
	// 这里开启ws监听
	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	count, err := a.task.BeforeWebSocketCheck(a.ctx, a.task.GetTaskID(a.ctx))
	if count == 0 {
		a.logger.Info(a.ctx, "count is 0")
		return nil
	}
	go func() {
		defer close(done)
		var servData = &ServData{TaskID: a.task.GetTaskID(a.ctx), HttpEndPoint: a.httpEndpoint, ProcessData: make([]string, 0)}
		for {
			var message []byte
			var messageType int
			messageType, message, err = a.ws.ReadMessage()
			if err != nil {
				a.logger.Info(a.ctx, "Error reading message from ws server:", err)
				a.errChan <- err
				return
			}
			switch messageType {
			case websocket.BinaryMessage:
				a.logger.Info(a.ctx, "binData, message len:", len(message))
				err = a.task.ParseBinData(a.ctx, message)
				if err != nil {
					a.logger.Info(a.ctx, "Error parse message from ws server ParseBinData:", err)
					a.errChan <- err
					continue
				}
			case websocket.TextMessage:
				var messageData = &MessageData{}
				err = json.Unmarshal(message, messageData)
				if err != nil {
					a.logger.Info(a.ctx, "Error decoding message from ws server:", err)
					a.errChan <- err
					return
				}
				if _, ok := SupportTypeMap[messageData.Type]; !ok {
					continue
				}
				err = json.Unmarshal(message, servData)
				if err != nil {
					a.logger.Info(a.ctx, "Error decoding message from ws server:", err)
					a.errChan <- err
					return
				}
				a.logger.Info(a.ctx, "message info:", string(message))
				var terminal bool

				switch messageData.Type {
				case Executed:
					if servData.Data != nil && servData.Data.Output != nil && len(servData.Data.Output.Images) > 0 {
						terminal, err = a.task.ExecutedImages(a.ctx, servData)
					}
					if servData.Data != nil && servData.Data.Output != nil && len(servData.Data.Output.Text) > 0 {
						terminal, err = a.task.ExecutedText(a.ctx, servData)
					}
					if servData.Data != nil && servData.Data.Output != nil && len(servData.Data.Output.Tags) > 0 {
						terminal = false
					}
					a.logger.Info(a.ctx, "other content terminal", servData)
				case Executing:
					fmt.Println("data.Node:", servData.Data.Node)
					if servData.Data.Node == "" {
						fmt.Println("data is null")
						// idle模式
						var outputMap map[string]*comfyui.HistoryData
						outputMap, err = comfyui.History(a.ctx, servData.HttpEndPoint)
						if err != nil {
							a.logger.Info(a.ctx, "history go error:", err)
							terminal = true
							done <- struct{}{}
							return
						}
						promptID = servData.Data.PromptID
						if outputMap[promptID] != nil {
							if outputMap[promptID] != nil && outputMap[promptID].Outputs != nil && len(outputMap[promptID].Outputs) == 1 {
								outputData := outputMap[promptID].Outputs
								for outputNodeID, output := range outputData {
									resultData := &OutputData{Images: make([]*comfyui.ImageInfo, 0), Text: make([]string, 0)}
									continueType := "images"
									if output.Images != nil && len(output.Images) > 0 {
										resultData.Images = output.Images
									}
									if output.Text != nil && len(output.Text) > 0 {
										resultData.Text = output.Text
										continueType = "text"
									}
									continueData := &ServData{
										Type: Executed,
										Data: &ServInnerData{
											Node:        outputNodeID,
											PromptID:    promptID,
											ExecMessage: "",
											Output:      resultData,
										},
										TaskID:         servData.TaskID,
										HttpEndPoint:   servData.HttpEndPoint,
										ProcessData:    servData.ProcessData,
										PreviewNodeMap: servData.PreviewNodeMap,
									}
									switch continueType {
									case "images":
										terminal, err = a.task.ExecutedImages(a.ctx, continueData)
									case "text":
										terminal, err = a.task.ExecutedText(a.ctx, continueData)
									}
									if err != nil {
										a.logger.Info(a.ctx, "Executing continue happend error:", err)
										terminal = true
									}
								}
							}
						}
					} else {
						terminal, err = a.task.Executing(a.ctx, servData)
					}
				case ExecutionStart:
					terminal, err = a.task.ExecutionStart(a.ctx, servData)
				case ExecutionError:
					terminal, err = a.task.ExecutionError(a.ctx, servData)
				default:
					a.logger.Info(a.ctx, "message data type is not supported", messageData.Type)
				}
				if err != nil {
					a.logger.Info(a.ctx, "Error executing message hook from client:", err)
					a.errChan <- err
					return
				}
				if terminal {
					a.logger.Info(a.ctx, "==========done operation")
					done <- struct{}{}
					return
				}
			}
		}
	}()
	select {
	case <-done:
		a.logger.Info(a.ctx, "case done triggered....")
		a.logger.Info(a.ctx, "Received done signal, closing connection...")
		_ = a.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		time.Sleep(time.Second) // 等待服务器处理关闭请求
		return nil
	case <-interrupt:
		a.logger.Info(a.ctx, "Received interrupt signal, closing connection...")
		_ = a.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		time.Sleep(time.Second) // 等待服务器处理关闭请求
		return nil
	}
}

func (a *aigc) queuePrompt() (promptID string, err error) {
	a.logger.Info(a.ctx, "start queue prompt")
	// 可能会有 extra_data
	// 预期字段: prompt
	promptDataMap := a.task.GetPrompt(a.ctx)

	if len(promptDataMap) == 0 {
		err = errors.New("prompt data map is empty")
		return
	}
	// 如果没有client_id 的话 给带上去 client_id 是一个区分websocket的标识，在 /prompt接口带上的 和 websocket连接带上的 clientId 就可以区分到,prompt的处理信息
	if _, ok := promptDataMap["client_id"]; !ok {
		promptDataMap["client_id"] = a.clientID
	}

	promptBytes, err := json.Marshal(promptDataMap)
	if err != nil {
		return "", err
	}
	_ = a.task.RecordParams(a.ctx, string(promptBytes))
	a.logger.Info(a.ctx, "queue prompt endpoint:", a.httpEndpoint)
	promptID, err = comfyui.QueuePrompt(a.ctx, a.httpEndpoint, string(promptBytes))
	if err != nil {
		return
	}
	return promptID, nil
}

// StartWebSocketConn wsEndpoint like: ws://127.0.0.1:8443/ws?clientId=1234456
func (a *aigc) connectWebSocketConn() error {
	c, _, err := websocket.DefaultDialer.Dial(a.wsEndPoint, nil)
	a.logger.Info(a.ctx, "dial ws:", a.wsEndPoint)
	if err != nil {
		a.close()
		a.logger.Info(a.ctx, "web socket conn error:", err)
		return nil
	}
	a.ws = c
	a.logger.Info(a.ctx, "StartWebSocketConn normal end")
	return nil
}
