package mysql

import "context"

var sqlPreprocessor func(ctx context.Context, query string) (string, error)

// SetSQLPreprocessor set preprocessor
func SetSQLPreprocessor(processor func(ctx context.Context, query string) (string, error)) {
	sqlPreprocessor = processor
}

func dummyPreprocessor(ctx context.Context, query string) (string, error) {
	return query, nil
}

func init() {
	SetSQLPreprocessor(dummyPreprocessor)
}
