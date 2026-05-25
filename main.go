package main

import (
	"context"
	"series-batch-go/internal/batch"
	"series-batch-go/internal/book"
	"series-batch-go/internal/config"
	"series-batch-go/internal/config/db"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/schedule"

	"google.golang.org/genai"
)

func main() {
	ctx, conf := context.Background(), config.Read()
	log.NewLogger(conf.Logger)

	postgres := db.NewGorm(conf.DB)
	defer func() {
		sql, _ := postgres.DB()
		_ = sql.Close()
	}()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: conf.Gemini.APIKey,
	})
	if err != nil {
		panic(err)
	}

	gemini := batch.NewGemini(conf.Gemini.APIKey, client, batch.ModelGemini3_5Flash)

	//bookRepository := book.NewGormRepository(postgres)
	//batchRepository := batch.NewGormRepository(postgres)
	//
	//reader := batch.NewSeriesClassifierReader(bookRepository)
	//writer := batch.NewSeriesClassifierWriter(gemini, batchRepository)
	//writer.GenerateDisplayName = func() string {
	//	uid, _ := uuid.NewRandom()
	//	return uid.String()
	//}
	//
	//jobInstance, err := schedule.NewJobBuilder[*book.Book]("SERIES_NORMALIZE").
	//	WithReader(reader).
	//	WithWriter(writer).
	//	Build()
	//if err != nil {
	//	panic(err)
	//}

	//batchRepository := batch.NewGormRepository(postgres)
	//reader := batch.NewSeriesMonitorReader(gemini, batchRepository)
	//writer := batch.NewSeriesMonitorWriter(batchRepository)
	//
	//jobInstance, err := schedule.NewJobBuilder[*batch.Batch]("BATCH_MONITOR").
	//	WithReader(reader).
	//	WithWriter(writer).
	//	Build()
	//if err != nil {
	//	panic(err)
	//}

	bookRepository := book.NewGormRepository(postgres)
	seriesRepository := book.NewSeriesGormRepository(postgres)
	batchRepository := batch.NewGormRepository(postgres)
	reader := batch.NewSeriesMapperReader(gemini, bookRepository, seriesRepository, batchRepository)
	writer := batch.NewSeriesMappingWriter(batchRepository, bookRepository, seriesRepository)

	jobInstance, err := schedule.NewJobBuilder[*batch.Mapped]("SERIES_MAPPING").
		WithReader(reader).
		WithWriter(writer).
		Build()
	if err != nil {
		panic(err)
	}

	instService := batch.NewJobService(batch.NewGormJobRepository(postgres))
	executor := schedule.NewExecutor(jobInstance, instService)

	parameter := make(map[string]string)
	if err = executor.Run(ctx, parameter); err != nil {
		panic(err)
	}
}
