package pkgLogger

import "log/slog"

// Err возвращает атрибут для ошибки.
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}