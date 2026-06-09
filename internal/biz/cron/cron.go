package cron

import (
	"context"
	"fmt"
	"time"

	"eino-stock/internal/data"
	croninfra "eino-stock/internal/infrastructure/cron"
)

type CronUsecase struct {
	repo  *data.CronTaskRepo
	sched *croninfra.Scheduler
}

func NewCronUsecase(repo *data.CronTaskRepo, sched *croninfra.Scheduler) *CronUsecase {
	uc := &CronUsecase{repo: repo, sched: sched}
	tasks, _ := repo.ListEnabled(context.Background())
	for _, t := range tasks {
		task := t
		uc.sched.Add(task.ID, task.Expression, func() { uc.executeTask(context.Background(), &task) })
	}
	return uc
}

func (uc *CronUsecase) List(ctx context.Context) ([]data.CronTask, error) {
	return uc.repo.List(ctx)
}

func (uc *CronUsecase) Create(ctx context.Context, t *data.CronTask) error {
	if err := uc.repo.Create(ctx, t); err != nil {
		return err
	}
	if t.Enable {
		task := *t
		uc.sched.Add(task.ID, task.Expression, func() { uc.executeTask(context.Background(), &task) })
	}
	return nil
}

func (uc *CronUsecase) Update(ctx context.Context, t *data.CronTask) error {
	if err := uc.repo.Update(ctx, t); err != nil {
		return err
	}
	uc.sched.Remove(t.ID)
	if t.Enable {
		task := *t
		uc.sched.Add(task.ID, task.Expression, func() { uc.executeTask(context.Background(), &task) })
	}
	return nil
}

func (uc *CronUsecase) Delete(ctx context.Context, id uint) error {
	uc.sched.Remove(id)
	return uc.repo.Delete(ctx, id)
}

func (uc *CronUsecase) Enable(ctx context.Context, id uint, enable bool) error {
	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	uc.sched.Remove(id)
	if enable {
		uc.sched.Add(task.ID, task.Expression, func() { uc.executeTask(context.Background(), task) })
	}
	return nil
}

func (uc *CronUsecase) ExecuteNow(ctx context.Context, id uint) error {
	task, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.executeTask(ctx, task)
}

func (uc *CronUsecase) executeTask(ctx context.Context, task *data.CronTask) error {
	now := time.Now()
	runResult := fmt.Sprintf("执行完成: %s", task.TaskType)
	return uc.repo.UpdateRunInfo(ctx, task.ID, now, runResult)
}
