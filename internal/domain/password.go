package domain

type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) error
}
