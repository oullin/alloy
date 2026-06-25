package encryption

// EncrypterContract encrypts and decrypts values.
type EncrypterContract interface {
	Encrypt(value any, serialize bool) (string, error)
	Decrypt(payload string, unserialize bool) (any, error)
	GetKey() string
	GetAllKeys() []string
	GetPreviousKeys() []string
}

// StringEncrypterContract encrypts and decrypts strings without serialization.
type StringEncrypterContract interface {
	EncryptString(value string) (string, error)
	DecryptString(payload string) (string, error)
}
