package cookie

// Encrypter encrypts and decrypts cookie values.
type Encrypter interface {
	Encrypt(value string) (string, error)
	Decrypt(value string) (string, error)
}
