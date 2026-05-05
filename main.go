package main

import (
	"context"
	"series-batch-go/internal/batch"
	"series-batch-go/internal/book"
	"series-batch-go/internal/config"
	"series-batch-go/internal/config/db"
	"series-batch-go/internal/config/log"
	"series-batch-go/internal/gemini"
	"series-batch-go/internal/job"
	"series-batch-go/internal/job/mapper"

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

	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: conf.Gemini.APIKey,
	})
	if err != nil {
		panic(err)
	}

	geminiClient := gemini.NewClient(genaiClient, gemini.ModelGemini3FlashPreview)

	//btr := batch.NewRepository(postgres)
	//reader := monitor.NewReader(btr)
	//processor := monitor.NewProcessor(geminiClient)
	//writer := monitor.NewWriter(btr)
	//j := job.NewBuilder[*monitor.ReadItem, *monitor.ProcessItem]().
	//	WithReader(reader).
	//	WithProcessor(processor).
	//	WithWriter(writer).
	//	WithChunkSize(100).
	//	Build()
	//err = j.Run(ctx, map[string]string{})
	//if err != nil {
	//	panic(err)
	//}

	//br := book.NewRepository(postgres)
	//btr := batch.NewRepository(postgres)
	//reader := classifier.NewReader(btr, br)
	//processor := &classifier.IdentifyProcessor{}
	//writer := classifier.NewWriter(geminiClient, btr)
	//
	//j := job.NewBuilder[*book.Book, *book.Book]().
	//	WithReader(reader).
	//	WithProcessor(processor).
	//	WithWriter(writer).
	//	WithChunkSize(100).
	//	Build()
	//
	//err = j.Run(ctx, map[string]string{})
	//if err != nil {
	//	panic(err)
	//}

	br := book.NewRepository(postgres)
	sr := book.NewSeriesRepository(postgres)
	btr := batch.NewRepository(postgres)

	reader := mapper.NewReader(geminiClient, btr, br, sr)
	processor := mapper.NewProcessor(br, sr)
	writer := mapper.NewWriter(btr, br, sr)

	j := job.NewBuilder[*mapper.ReadItem, *mapper.ProcessItem]().
		WithReader(reader).
		WithProcessor(processor).
		WithWriter(writer).
		WithChunkSize(1).
		Build()

	err = j.Run(ctx, map[string]string{
		"batch_size": "100",
	})
	if err != nil {
		panic(err)
	}
}
