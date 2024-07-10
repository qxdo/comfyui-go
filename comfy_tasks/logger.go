package comfy_tasks

import "context"

type Logger interface {
	Info(ctx context.Context, data ...interface{})
}
