package comfy_tasks

import (
	"context"
	"errors"
	"time"
)

type AigcTaskProcessor struct {
	ctx         context.Context
	task        AigcTask
	taskHandler *aigc
	logger      Logger
}

func NewAigcTaskProcessor(ctx context.Context, httpEndpoint, wsEndpoint string, task AigcTask, logger Logger) *AigcTaskProcessor {
	aTask := New(ctx, httpEndpoint, wsEndpoint, task, logger)
	return &AigcTaskProcessor{
		ctx:         ctx,
		task:        task,
		taskHandler: aTask,
		logger:      logger,
	}
}

func (a *AigcTaskProcessor) Start() error {
	svr := a.taskHandler
	defer svr.close()
	count, err := a.task.BeforeWebSocketCheck(a.ctx, a.task.GetTaskID(a.ctx))
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	doneChannel := make(chan struct{}, 1)
	ticker := time.NewTicker(time.Duration(a.task.GetTaskTimeoutTickerTime()) * time.Second)
	defer ticker.Stop()
	go func() {
		err = svr.Start()
		if err != nil {
			a.logger.Info(a.ctx, "err close sig:", err)
			svr.errChan <- err
		}
		a.logger.Info(a.ctx, "ticker start goroutine end")
		doneChannel <- struct{}{}
	}()

	for {
		select {
		case <-ticker.C:
			a.logger.Info(a.ctx, "ticker.C normal close")
			svr.errChan <- errors.New("timeout ticker close server")
			return nil
		case err = <-svr.errChan:
			msg := "Start Error Handled"
			if err != nil {
				msg = err.Error()
			}
			if err = svr.task.TaskFailed(svr.ctx, "process_error_handled", msg); err != nil {
				a.logger.Info(a.ctx, "task failed by errChannel", err)
			}
			return nil
		case <-doneChannel:
			a.logger.Info(a.ctx, "normal end")
			return nil
		}
	}
}
