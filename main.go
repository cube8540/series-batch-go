package main

import (
	"context"
	"flag"
	"fmt"
	"series-batch-go/internal/batch"
	"series-batch-go/internal/book"
	"series-batch-go/internal/config"
	"series-batch-go/internal/config/db"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/schedule"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

type batchName string

const (
	batchNameSeriesNormalize batchName = "SERIES_NORMALIZE"
	batchNameBatchMonitor    batchName = "BATCH_MONITOR"
	batchNameSeriesMapping   batchName = "SERIES_MAPPING"
	batchNameCleanup         batchName = "CLEANUP"
	batchNameRecovery        batchName = "RECOVERY"
)

type Task interface {
	Run(context.Context, schedule.JobParameter) error
}

func main() {
	inputBatchName := flag.String("batch", "", "실행시킬 배치명")
	flag.Parse()

	name, err := newBatchName(*inputBatchName)
	if err != nil {
		panic(err)
	}

	ctx, conf := context.Background(), config.Read()
	log.NewLogger(conf.Logger)

	postgres := db.NewGorm(conf.DB)
	defer func() {
		sql, _ := postgres.DB()
		_ = sql.Close()
	}()

	ai, err := newAI(ctx, conf.Gemini.APIKey)
	if err != nil {
		panic(err)
	}
	instService := batch.NewJobService(batch.NewGormJobRepository(postgres))

	var task Task
	switch name {
	case batchNameSeriesNormalize:
		bookRepository := book.NewGormRepository(postgres)
		batchRepository := batch.NewGormRepository(postgres)

		reader := batch.NewSeriesClassifierReader(bookRepository)
		writer := batch.NewSeriesClassifierWriter(ai, batchRepository)
		writer.GenerateDisplayName = func() string {
			uid, _ := uuid.NewRandom()
			return uid.String()
		}

		inst, err := schedule.NewJobBuilder[*book.Book](string(name)).
			WithReader(reader).
			WithWriter(writer).
			Build()
		if err != nil {
			panic(err)
		}
		task = schedule.NewExecutor(inst, instService)
	case batchNameBatchMonitor:
		batchRepository := batch.NewGormRepository(postgres)
		reader := batch.NewSeriesMonitorReader(ai, batchRepository)
		writer := batch.NewSeriesMonitorWriter(batchRepository)

		inst, err := schedule.NewJobBuilder[*batch.Batch](string(name)).
			WithReader(reader).
			WithWriter(writer).
			Build()
		if err != nil {
			panic(err)
		}
		task = schedule.NewExecutor(inst, instService)
	case batchNameSeriesMapping:
		bookRepository := book.NewGormRepository(postgres)
		seriesRepository := book.NewSeriesGormRepository(postgres)
		batchRepository := batch.NewGormRepository(postgres)

		reader := batch.NewSeriesMapperReader(ai, bookRepository, seriesRepository, batchRepository)
		writer := batch.NewSeriesMappingWriter(batchRepository, bookRepository, seriesRepository)

		inst, err := schedule.NewJobBuilder[*batch.Mapped](string(name)).
			WithReader(reader).
			WithWriter(writer).
			Build()
		if err != nil {
			panic(err)
		}
		task = schedule.NewExecutor(inst, instService)
	case batchNameCleanup:
		batchRepository := batch.NewGormRepository(postgres)

		reader := batch.NewCleanupBatchReader(batchRepository)
		writer := batch.NewCleanupBatchWriter(batchRepository)

		inst, err := schedule.NewJobBuilder[*batch.Batch](string(name)).
			WithReader(reader).
			WithWriter(writer).
			Build()
		if err != nil {
			panic(err)
		}
		task = schedule.NewExecutor(inst, instService)
	case batchNameRecovery:
		batchRepository := batch.NewGormRepository(postgres)
		bookRepository := book.NewGormRepository(postgres)

		reader := batch.NewRecoveryBatchReader(batchRepository, bookRepository)
		writer := batch.NewSeriesClassifierWriter(ai, batchRepository)
		writer.GenerateDisplayName = func() string {
			uid, _ := uuid.NewRandom()
			return uid.String()
		}

		inst, err := schedule.NewJobBuilder[*book.Book](string(name)).
			WithReader(reader).
			WithWriter(writer).
			WithChunkSize(1000).
			Build()
		if err != nil {
			panic(err)
		}

		executor := schedule.NewExecutor(inst, instService)
		executor.AddListener(batch.NewRecoveryCompletedEventListener(batchRepository))

		task = executor
	}

	params := make(map[string]string)
	if err = task.Run(ctx, params); err != nil {
		panic(err)
	}
}

func newBatchName(name string) (batchName, error) {
	name = strings.ToUpper(name)
	switch name {
	case "SERIES_NORMALIZE":
		return batchNameSeriesNormalize, nil
	case "BATCH_MONITOR":
		return batchNameBatchMonitor, nil
	case "SERIES_MAPPING":
		return batchNameSeriesMapping, nil
	case "CLEANUP":
		return batchNameCleanup, nil
	case "RECOVERY":
		return batchNameRecovery, nil
	default:
		return "", fmt.Errorf("unknown batch name: %s", name)
	}
}

func newAI(ctx context.Context, key string) (batch.AI, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return batch.NewGemini(key, client, batch.ModelGemini3_1FlashLite), nil
}
