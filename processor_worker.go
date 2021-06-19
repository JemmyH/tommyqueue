package tommyqueue

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/jemmyh/tommyqueue/internal/broker"
	"github.com/jemmyh/tommyqueue/internal/structs"
)

var (
	// ProcessHandler 的返回值，表示当前 task 需要 retry or archive
	ErrSkipRetry = errors.New("shoule retry later")
)

type processorWorker struct {
	broker broker.CtxBroker
	logger Logger

	taskHandler   ProcessHandler
	errHandler    ErrorHandler
	retryDelayFun RetryDelayFunc

	done chan struct{}

	// quit is closed when the processor shutdowns.
	quit chan struct{}

	// only get this token a task can be processed
	token chan struct{}

	// 任务 timeout
	timeout chan struct{}

	// 队列名称，processor 根据队列名称从 broker 中拉取任务
	prioQueenNames []string
}

func (p *processorWorker) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-p.done:
				p.logger.Debug("processor done")
				return
			default:
				p.process(ctx)
			}
		}
	}()
}

func (p *processorWorker) Stop(ctx context.Context) {
	panic("implement me")
}

func (p *processorWorker) process(ctx context.Context) {
	select {
	case <-p.quit:
		return
	case p.token <- struct{}{}:
		// acquire token
		queenNames := p.getPrioQueenName(ctx)
		msg, ddl, err := p.broker.Dequeue(ctx, queenNames...)
		if err != nil {
			if err == broker.ErrNoReadyTask {
				p.logger.Debug("no ready tasks in queens")
				time.Sleep(time.Second)
				<-p.token // release token
				return
			}
			p.logger.Error("dequeen error: " + err.Error())
			<-p.token
			return
		}
		// start a goroutine to process task
		go func() {
			defer func() {
				<-p.token
			}()
			newCtx, cancel := newTaskContext(ctx, msg, ddl)
			defer cancel()

			// before starting a worker goroutine, check whether context has been canceled
			select {
			case <-newCtx.Done():
				p.retryOrKill(ctx, msg, newCtx.Err())
				return
			default:
			}

			resChan := make(chan error, 1)
			go func() {
				resChan <- p.processTask(ctx, NewTask(msg.Type.String(), msg.Payload))
			}()

			select {
			case <-p.timeout:
				p.logger.Error("timeout for task" + msg.ID.String() + ", repush this task to queue and quit worker...")
				p.reEnQueue(ctx, msg)
				return
			case <-newCtx.Done():
				p.retryOrKill(ctx, msg, newCtx.Err())
				return
			case resErr := <-resChan:
				/*
					有三种情形：
					1) Done     -> Removes the message from Active
					2) Retry    -> Removes the message from Active & Adds the message to Retry
					3) Archive  -> Removes the message from Active & Adds the message to archive
				*/
				if resErr != nil {
					p.retryOrKill(ctx, msg, resErr)
					return
				}
				p.markAsDone(ctx, msg)
			}
		}()
	}
}

// processTask 调用 worker 的 ProcessTask，执行用户处理逻辑
func (p *processorWorker) processTask(ctx context.Context, task *Task) (err error) {
	defer func() {
		// handle panic here
		if x := recover(); x != nil {
			p.logger.Errorf("recovering from panic. Stack details:\n%s", string(debug.Stack()))
			_, file, line, ok := runtime.Caller(1) // skip the first frame (panic itself)
			if ok && strings.Contains(file, "runtime/") {
				// The panic came from the runtime, most likely due to incorrect
				// map/slice usage. The parent frame should have the real trigger.
				_, file, line, ok = runtime.Caller(2)
			}

			// Include the file and line number info in the error, if runtime.Caller returned ok.
			if ok {
				err = fmt.Errorf("panic [%s:%d]: %v", file, line, x)
			} else {
				err = fmt.Errorf("panic: %v", x)
			}
		}
	}()
	return p.taskHandler.ProcessTask(ctx, task)
}

func (p *processorWorker) retryOrKill(ctx context.Context, msg *structs.TaskMessage, e error) {
	if p.errHandler != nil {
		p.errHandler.HandleError(ctx, NewTask(msg.Type.String(), msg.Payload), e)
	}
	if msg.CanRetry <= 0 || errors.Is(e, ErrSkipRetry) {
		// this task has tries enough times, just archive it.
		p.logger.Warnf("task %s exhausted, archive it", msg.ID)
		p.archiveTask(ctx, msg, e)
	} else {
		p.retryTask(ctx, msg, e)
	}
}

func (p *processorWorker) archiveTask(ctx context.Context, msg *structs.TaskMessage, e error) {
	if err := p.broker.Archive(ctx, msg, e.Error()); err != nil {
		// TODO: retry again?
	}
}

func (p *processorWorker) retryTask(ctx context.Context, msg *structs.TaskMessage, e error) {
	delay := p.retryDelayFun(msg.CanRetry, e, NewTask(msg.Type.String(), msg.Payload))
	retryAt := time.Now().Add(delay)
	if err := p.broker.Retry(ctx, msg, retryAt, e.Error()); err != nil {
		// TODO: retry again?
	}
}

func (p *processorWorker) markAsDone(ctx context.Context, msg *structs.TaskMessage) {
	if err := p.broker.Done(ctx, msg); err != nil {
		// TODO: retry again?
	}
}

func (p *processorWorker) reEnQueue(ctx context.Context, msg *structs.TaskMessage) {
	err := p.broker.Requeue(ctx, msg)
	if err != nil {
		p.logger.Errorf("push task %s to queue error: %v", msg.ID, err)
		return
	}
	p.logger.Infof("push task %s to queue ok", msg.ID)
}

func (p *processorWorker) getPrioQueenName(ctx context.Context) []string {
	return p.prioQueenNames
}
