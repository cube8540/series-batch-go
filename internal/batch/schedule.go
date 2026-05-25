package batch

import (
	"context"
	"errors"
	"fmt"
	"series-batch-go/internal/schedule"
)

type JobService struct {
	repo JobRepository
}

func NewJobService(repo JobRepository) *JobService {
	return &JobService{repo: repo}
}

func (serv JobService) Get(ctx context.Context, name string) (schedule.Instance, error) {
	if name == "" {
		return schedule.Instance{}, errors.New("name cannot be empty")
	}

	job, err := serv.repo.Get(ctx, name)
	if err != nil {
		return schedule.Instance{}, fmt.Errorf("failed to get job instance: %w", err)
	}

	instance := schedule.Instance{
		Name:  job.Name,
		State: schedule.JobStatus(job.State),
	}
	return instance, nil
}

func (serv JobService) Update(ctx context.Context, inst schedule.Instance) error {
	return serv.repo.Transaction(ctx, func(repo JobRepository) error {
		job, err := repo.GetForUpdate(ctx, inst.Name)
		if err != nil {
			return err
		}

		job.State = string(inst.State)
		return repo.Update(ctx, job)
	})
}
