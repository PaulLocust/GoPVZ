package pkgValidator

// Validator — интерфейс для валидации любой структуры.
type Validator interface {
	Validate() error
}
