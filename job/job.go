package job

import "context"

type Job interface {
	Execute(context.Context) error
	GetJobName() string
}
