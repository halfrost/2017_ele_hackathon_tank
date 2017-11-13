package utils

// RetryFunc is a type which used for Retry.
type RetryFunc func() error

// Retry rerun the fn maxRetry times as long as error is not nil.
func Retry(fn RetryFunc, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		if err = fn(); err == nil {
			return nil
		}
	}
	return err
}
