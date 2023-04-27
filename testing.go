package got

//go:generate mockgen -destination=testing_mock.go -package=got . T

type T interface {
	Helper()
	Fatal(...any)
}
