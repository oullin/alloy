package encryption

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestEncryption(t *testing.T) {
	t.Parallel()

	for _, c := range []Cipher{AES128CBC, AES256CBC, AES128GCM, AES256GCM} {
		t.Run(string(c), func(t *testing.T) {
			t.Parallel()

			key := mustGenerateKey(t, c)
			enc := mustNewEncrypter(t, key, c)

			encrypted, err := enc.EncryptString("foo")

			if err != nil {
				t.Fatal(err)
			}

			result, err := enc.DecryptString(encrypted)

			if err != nil {
				t.Fatal(err)
			}

			if result != "foo" {
				t.Fatalf("expected foo, got %s", result)
			}
		})
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)

	encrypted, err := enc.EncryptString("")

	if err != nil {
		t.Fatal(err)
	}

	result, err := enc.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestEncryptDecryptLongString(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	long := strings.Repeat("a", 10000)

	encrypted, err := enc.EncryptString(long)

	if err != nil {
		t.Fatal(err)
	}

	result, err := enc.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != long {
		t.Fatal("long string round-trip failed")
	}
}

func TestEncryptDecryptWithSerialization(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)

	data := map[string]any{"foo": "bar", "baz": float64(42)}
	encrypted, err := enc.Encrypt(data, true)

	if err != nil {
		t.Fatal(err)
	}

	result, err := enc.Decrypt(encrypted, true)

	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[string]any)

	if !ok {
		t.Fatal("expected map")
	}

	if m["foo"] != "bar" {
		t.Fatalf("expected bar, got %v", m["foo"])
	}

	if m["baz"] != float64(42) {
		t.Fatalf("expected 42, got %v", m["baz"])
	}
}

func TestRawStringEncryption(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES128CBC), AES128CBC)

	encrypted, err := enc.EncryptString("hello world")

	if err != nil {
		t.Fatal(err)
	}

	result, err := enc.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "hello world" {
		t.Fatalf("expected hello world, got %s", result)
	}
}

func TestRawStringEncryptionWithPreviousKeys(t *testing.T) {
	t.Parallel()

	keyOld := mustGenerateKey(t, AES256CBC)
	keyNew := mustGenerateKey(t, AES256CBC)

	encOld := mustNewEncrypter(t, keyOld, AES256CBC)
	encrypted, err := encOld.EncryptString("secret")

	if err != nil {
		t.Fatal(err)
	}

	encNew := mustNewEncrypter(t, keyNew, AES256CBC)
	encNew.PreviousKeys([][]byte{keyOld})

	result, err := encNew.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "secret" {
		t.Fatalf("expected secret, got %s", result)
	}
}

func TestMACValidationPerKey(t *testing.T) {
	t.Parallel()

	keyA := mustGenerateKey(t, AES256CBC)
	keyB := mustGenerateKey(t, AES256CBC)
	keyC := mustGenerateKey(t, AES256CBC)

	encA := mustNewEncrypter(t, keyA, AES256CBC)
	encrypted, err := encA.EncryptString("data")

	if err != nil {
		t.Fatal(err)
	}

	encC := mustNewEncrypter(t, keyC, AES256CBC)
	encC.PreviousKeys([][]byte{keyB, keyA})

	result, err := encC.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "data" {
		t.Fatalf("expected data, got %s", result)
	}
}

func TestEncryptionUsingBase64EncodedKey(t *testing.T) {
	t.Parallel()

	raw := mustGenerateKey(t, AES256CBC)
	encoded := "base64:" + base64.StdEncoding.EncodeToString(raw)

	key, err := ParseKey(encoded)

	if err != nil {
		t.Fatal(err)
	}

	enc := mustNewEncrypter(t, key, AES256CBC)

	encrypted, err := enc.EncryptString("test")

	if err != nil {
		t.Fatal(err)
	}

	result, err := enc.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "test" {
		t.Fatalf("expected test, got %s", result)
	}
}

func TestWithCustomCipher(t *testing.T) {
	t.Parallel()

	key := mustGenerateKey(t, AES256GCM)
	enc := mustNewEncrypter(t, key, AES256GCM)

	encrypted, err := enc.EncryptString("gcm-test")

	if err != nil {
		t.Fatal(err)
	}

	result, err := enc.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "gcm-test" {
		t.Fatalf("expected gcm-test, got %s", result)
	}
}

func TestCipherNamesCanBeMixedCase(t *testing.T) {
	t.Parallel()

	cases := []string{"AES-256-CBC", "Aes-256-Cbc", "aes-256-cbc", "AES-256-cbc"}

	for _, name := range cases {
		c, err := ParseCipher(name)

		if err != nil {
			t.Fatalf("ParseCipher(%q) failed: %v", name, err)
		}

		if c != AES256CBC {
			t.Fatalf("expected %s, got %s", AES256CBC, c)
		}
	}
}

func TestSupportedMethodAcceptsAnyCasing(t *testing.T) {
	t.Parallel()

	key := mustGenerateKey(t, AES128CBC)
	c, _ := ParseCipher("AES-128-CBC")

	if !Supported(key, c) {
		t.Fatal("expected Supported to return true")
	}
}

func TestAEADCipherIncludesTag(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256GCM), AES256GCM)
	encrypted, err := enc.EncryptString("tagged")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)

	if p.Tag == "" {
		t.Fatal("expected tag in AEAD payload")
	}

	if p.MAC != "" {
		t.Fatal("expected no mac in AEAD payload")
	}
}

func TestAEADTagMustBeFullLength(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256GCM), AES256GCM)
	encrypted, err := enc.EncryptString("tagged")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	tagBytes, _ := base64.StdEncoding.DecodeString(p.Tag)
	p.Tag = base64.StdEncoding.EncodeToString(tagBytes[:len(tagBytes)-1])
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if err == nil {
		t.Fatal("expected error for truncated tag")
	}
}

func TestAEADTagCantBeModified(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256GCM), AES256GCM)
	encrypted, err := enc.EncryptString("tagged")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	tagBytes, _ := base64.StdEncoding.DecodeString(p.Tag)
	tagBytes[0] ^= 0xff
	p.Tag = base64.StdEncoding.EncodeToString(tagBytes)
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if err == nil {
		t.Fatal("expected error for modified tag")
	}
}

func TestNonAEADCipherIncludesMAC(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("mac-test")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)

	if p.MAC == "" {
		t.Fatal("expected mac in CBC payload")
	}

	if p.Tag != "" {
		t.Fatal("expected no tag in CBC payload")
	}
}

func TestDoNotAllowLongerKey(t *testing.T) {
	t.Parallel()

	key := make([]byte, 64)
	_, err := NewEncrypter(key, AES256CBC)

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestWithBadKeyLength(t *testing.T) {
	t.Parallel()

	key := make([]byte, 8)
	_, err := NewEncrypter(key, AES128CBC)

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestWithBadKeyLengthAlternativeCipher(t *testing.T) {
	t.Parallel()

	key := make([]byte, 16)
	_, err := NewEncrypter(key, AES256GCM)

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestWithUnsupportedCipher(t *testing.T) {
	t.Parallel()

	_, err := ParseCipher("aes-512-xyz")

	if !errors.Is(err, ErrUnsupportedCipher) {
		t.Fatalf("expected ErrUnsupportedCipher, got %v", err)
	}
}

func TestExceptionThrownWhenPayloadIsInvalid(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)

	cases := []string{"", "not-base64!", "aGVsbG8=", base64.StdEncoding.EncodeToString([]byte("{}"))}

	for _, c := range cases {
		_, err := enc.DecryptString(c)

		if err == nil {
			t.Fatalf("expected error for payload %q", c)
		}
	}
}

func TestDecryptionExceptionWhenUnexpectedTagIsAdded(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("test")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.Tag = base64.StdEncoding.EncodeToString([]byte("unexpected-tag!!"))
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if err == nil {
		t.Fatal("expected error for unexpected tag on CBC")
	}
}

func TestExceptionThrownWhenIVIsTooLong(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("test")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.IV = base64.StdEncoding.EncodeToString(make([]byte, 32))
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if err == nil {
		t.Fatal("expected error for oversized IV")
	}
}

func TestTamperedPayloadWillGetRejected(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("tamper-test")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	valBytes, _ := base64.StdEncoding.DecodeString(p.Value)
	valBytes[0] ^= 0xff
	p.Value = base64.StdEncoding.EncodeToString(valBytes)
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if err == nil {
		t.Fatal("expected error for tampered value")
	}
}

func TestExceptionThrownWithDifferentKey(t *testing.T) {
	t.Parallel()

	keyA := mustGenerateKey(t, AES256CBC)
	keyB := mustGenerateKey(t, AES256CBC)

	encA := mustNewEncrypter(t, keyA, AES256CBC)
	encrypted, err := encA.EncryptString("secret")

	if err != nil {
		t.Fatal(err)
	}

	encB := mustNewEncrypter(t, keyB, AES256CBC)
	_, err = encB.DecryptString(encrypted)

	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestAppearsEncryptedReturnsTrueForEncryptedValue(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("test")

	if err != nil {
		t.Fatal(err)
	}

	if !AppearsEncrypted(encrypted) {
		t.Fatal("expected AppearsEncrypted to return true")
	}
}

func TestAppearsEncryptedReturnsTrueForEncryptedArray(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.Encrypt([]string{"a", "b"}, true)

	if err != nil {
		t.Fatal(err)
	}

	if !AppearsEncrypted(encrypted) {
		t.Fatal("expected AppearsEncrypted to return true")
	}
}

func TestAppearsEncryptedReturnsFalseForPlainText(t *testing.T) {
	t.Parallel()

	if AppearsEncrypted("hello world") {
		t.Fatal("expected AppearsEncrypted to return false for plaintext")
	}
}

func TestAppearsEncryptedReturnsFalseForEmptyString(t *testing.T) {
	t.Parallel()

	if AppearsEncrypted("") {
		t.Fatal("expected AppearsEncrypted to return false for empty string")
	}
}

func TestGenerateKeyLength(t *testing.T) {
	t.Parallel()

	for _, c := range []Cipher{AES128CBC, AES256CBC, AES128GCM, AES256GCM} {
		key, err := GenerateKey(c)

		if err != nil {
			t.Fatal(err)
		}

		if len(key) != c.KeyLength() {
			t.Fatalf("expected key length %d for %s, got %d", c.KeyLength(), c, len(key))
		}
	}
}

func TestSupportedValidation(t *testing.T) {
	t.Parallel()

	key16 := make([]byte, 16)
	key32 := make([]byte, 32)

	if !Supported(key16, AES128CBC) {
		t.Fatal("16-byte key should be supported for AES-128-CBC")
	}

	if !Supported(key32, AES256CBC) {
		t.Fatal("32-byte key should be supported for AES-256-CBC")
	}

	if Supported(key16, AES256CBC) {
		t.Fatal("16-byte key should not be supported for AES-256-CBC")
	}

	if Supported(key32, AES128GCM) {
		t.Fatal("32-byte key should not be supported for AES-128-GCM")
	}
}

func TestGetKeyAndGetAllKeys(t *testing.T) {
	t.Parallel()

	key := mustGenerateKey(t, AES256CBC)
	prev := mustGenerateKey(t, AES256CBC)
	enc := mustNewEncrypter(t, key, AES256CBC)
	enc.PreviousKeys([][]byte{prev})

	if enc.GetKey() != base64.StdEncoding.EncodeToString(key) {
		t.Fatal("GetKey mismatch")
	}

	all := enc.GetAllKeys()

	if len(all) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(all))
	}

	if all[0] != base64.StdEncoding.EncodeToString(key) {
		t.Fatal("first key should be current")
	}

	if all[1] != base64.StdEncoding.EncodeToString(prev) {
		t.Fatal("second key should be previous")
	}

	prevKeys := enc.GetPreviousKeys()

	if len(prevKeys) != 1 {
		t.Fatalf("expected 1 previous key, got %d", len(prevKeys))
	}
}

func TestNewEncrypterCopiesKey(t *testing.T) {
	t.Parallel()

	key := mustGenerateKey(t, AES256CBC)
	expected := base64.StdEncoding.EncodeToString(key)

	enc := mustNewEncrypter(t, key, AES256CBC)
	key[0] ^= 0xff

	if enc.GetKey() != expected {
		t.Fatal("mutating caller-owned key should not change encrypter key")
	}
}

func TestPreviousKeysCopiesKeys(t *testing.T) {
	t.Parallel()

	keyOld := mustGenerateKey(t, AES256CBC)
	keyNew := mustGenerateKey(t, AES256CBC)

	encOld := mustNewEncrypter(t, keyOld, AES256CBC)
	ciphertext, err := encOld.EncryptString("legacy")

	if err != nil {
		t.Fatal(err)
	}

	encNew := mustNewEncrypter(t, keyNew, AES256CBC)
	encNew.PreviousKeys([][]byte{keyOld})
	keyOld[0] ^= 0xff

	plaintext, err := encNew.DecryptString(ciphertext)

	if err != nil {
		t.Fatal(err)
	}

	if plaintext != "legacy" {
		t.Fatalf("expected legacy, got %s", plaintext)
	}
}

func TestEncryptDecryptRoundTripSerializedAllCiphers(t *testing.T) {
	t.Parallel()

	for _, c := range []Cipher{AES128CBC, AES256CBC, AES128GCM, AES256GCM} {
		t.Run(string(c), func(t *testing.T) {
			t.Parallel()

			enc := mustNewEncrypter(t, mustGenerateKey(t, c), c)

			encrypted, err := enc.Encrypt([]any{"a", float64(1)}, true)

			if err != nil {
				t.Fatal(err)
			}

			result, err := enc.Decrypt(encrypted, true)

			if err != nil {
				t.Fatal(err)
			}

			list, ok := result.([]any)

			if !ok || len(list) != 2 || list[0] != "a" || list[1] != float64(1) {
				t.Fatalf("round-trip mismatch: %#v", result)
			}
		})
	}
}

func TestWrongKeyFailsForAllCiphers(t *testing.T) {
	t.Parallel()

	for _, c := range []Cipher{AES128CBC, AES256CBC, AES128GCM, AES256GCM} {
		t.Run(string(c), func(t *testing.T) {
			t.Parallel()

			encA := mustNewEncrypter(t, mustGenerateKey(t, c), c)
			encB := mustNewEncrypter(t, mustGenerateKey(t, c), c)

			encrypted, err := encA.EncryptString("secret")

			if err != nil {
				t.Fatal(err)
			}

			_, err = encB.DecryptString(encrypted)

			if !errors.Is(err, ErrDecryptFailed) {
				t.Fatalf("expected ErrDecryptFailed, got %v", err)
			}
		})
	}
}

func TestPreviousKeysRotationGCM(t *testing.T) {
	t.Parallel()

	keyOld := mustGenerateKey(t, AES256GCM)
	keyNew := mustGenerateKey(t, AES256GCM)

	encOld := mustNewEncrypter(t, keyOld, AES256GCM)
	encrypted, err := encOld.EncryptString("rotated")

	if err != nil {
		t.Fatal(err)
	}

	encNew := mustNewEncrypter(t, keyNew, AES256GCM)
	encNew.PreviousKeys([][]byte{keyOld})

	result, err := encNew.DecryptString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	if result != "rotated" {
		t.Fatalf("expected rotated, got %s", result)
	}
}

func TestEncryptRejectsNonStringWithoutSerialization(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)

	_, err := enc.Encrypt(42, false)

	if !errors.Is(err, ErrEncryptFailed) {
		t.Fatalf("expected ErrEncryptFailed, got %v", err)
	}
}

func TestEncryptRejectsUnserializableValue(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)

	_, err := enc.Encrypt(make(chan int), true)

	if !errors.Is(err, ErrEncryptFailed) {
		t.Fatalf("expected ErrEncryptFailed, got %v", err)
	}
}

func TestDecryptUnserializeRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("not-json{")

	if err != nil {
		t.Fatal(err)
	}

	_, err = enc.Decrypt(encrypted, true)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestInvalidPayloadSentinel(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)

	cases := []string{"", "not-base64!", "aGVsbG8=", base64.StdEncoding.EncodeToString([]byte("{}"))}

	for _, c := range cases {
		_, err := enc.DecryptString(c)

		if !errors.Is(err, ErrInvalidPayload) {
			t.Fatalf("expected ErrInvalidPayload for %q, got %v", c, err)
		}
	}
}

func TestTruncatedPayloadIsInvalid(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("truncate-me")

	if err != nil {
		t.Fatal(err)
	}

	_, err = enc.DecryptString(encrypted[:len(encrypted)/2])

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestMissingMACIsInvalidPayload(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("no-mac")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.MAC = ""
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestMissingTagIsInvalidPayloadForGCM(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256GCM), AES256GCM)
	encrypted, err := enc.EncryptString("no-tag")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.Tag = ""
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestInvalidIVBase64IsInvalidPayload(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("bad-iv")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.IV = "%%%not-base64%%%"
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestTamperedMACIsRejected(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("mac-tamper")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	macBytes, _ := base64.StdEncoding.DecodeString(p.MAC)
	macBytes[0] ^= 0xff
	p.MAC = base64.StdEncoding.EncodeToString(macBytes)
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestInvalidMACBase64IsRejected(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	encrypted, err := enc.EncryptString("mac-garbage")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.MAC = "%%%not-base64%%%"
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestInvalidValueBase64IsRejected(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256GCM), AES256GCM)
	encrypted, err := enc.EncryptString("value-garbage")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.Value = "%%%not-base64%%%"
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestInvalidTagBase64IsRejected(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256GCM), AES256GCM)
	encrypted, err := enc.EncryptString("tag-garbage")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	p.Tag = "%%%not-base64%%%"
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestTamperedGCMCiphertextIsRejected(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES128GCM), AES128GCM)
	encrypted, err := enc.EncryptString("gcm-tamper")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	valBytes, _ := base64.StdEncoding.DecodeString(p.Value)
	valBytes[0] ^= 0xff
	p.Value = base64.StdEncoding.EncodeToString(valBytes)
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestTruncatedCiphertextWithForgedMACIsRejected(t *testing.T) {
	t.Parallel()

	key := mustGenerateKey(t, AES256CBC)
	enc := mustNewEncrypter(t, key, AES256CBC)
	encrypted, err := enc.EncryptString("block-size-check")

	if err != nil {
		t.Fatal(err)
	}

	p := decodePayload(t, encrypted)
	valBytes, _ := base64.StdEncoding.DecodeString(p.Value)
	p.Value = base64.StdEncoding.EncodeToString(valBytes[:len(valBytes)-1])
	p.MAC = base64.StdEncoding.EncodeToString(computeMAC(key, p.IV, p.Value))
	tampered := encodePayload(t, p)

	_, err = enc.DecryptString(tampered)

	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestPkcs7PadRoundTrip(t *testing.T) {
	t.Parallel()

	for _, size := range []int{0, 1, 15, 16, 17, 32} {
		data := make([]byte, size)

		for i := range data {
			data[i] = byte(i)
		}

		padded := pkcs7Pad(data, 16)

		if len(padded)%16 != 0 {
			t.Fatalf("padded length %d is not a block multiple", len(padded))
		}

		unpadded, err := pkcs7Unpad(padded)

		if err != nil {
			t.Fatal(err)
		}

		if len(unpadded) != size {
			t.Fatalf("expected %d bytes after unpad, got %d", size, len(unpadded))
		}
	}
}

func TestPkcs7UnpadRejectsInvalidPadding(t *testing.T) {
	t.Parallel()

	cases := map[string][]byte{
		"empty":              {},
		"zero padding":       {1, 2, 3, 0},
		"padding over block": append(make([]byte, 15), 17),
		"padding over data":  {5},
		"inconsistent bytes": {1, 2, 3, 3, 2, 3},
	}

	for name, data := range cases {
		if _, err := pkcs7Unpad(data); !errors.Is(err, ErrDecryptFailed) {
			t.Fatalf("%s: expected ErrDecryptFailed, got %v", name, err)
		}
	}
}

func TestPreviousKeysEmptyAndNilEntries(t *testing.T) {
	t.Parallel()

	enc := mustNewEncrypter(t, mustGenerateKey(t, AES256CBC), AES256CBC)
	enc.PreviousKeys(nil)

	if len(enc.GetPreviousKeys()) != 0 {
		t.Fatal("expected no previous keys")
	}

	if len(enc.GetAllKeys()) != 1 {
		t.Fatal("expected only the current key")
	}

	enc.PreviousKeys([][]byte{nil})

	prev := enc.GetPreviousKeys()

	if len(prev) != 1 || prev[0] != "" {
		t.Fatalf("expected single empty previous key, got %v", prev)
	}
}

// --- test helpers ---

func mustGenerateKey(t *testing.T, c Cipher) []byte {
	t.Helper()

	key, err := GenerateKey(c)

	if err != nil {
		t.Fatal(err)
	}

	return key
}

func mustNewEncrypter(t *testing.T, key []byte, c Cipher) *Encrypter {
	t.Helper()

	enc, err := NewEncrypter(key, c)

	if err != nil {
		t.Fatal(err)
	}

	return enc
}

func decodePayload(t *testing.T, encrypted string) payload {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(encrypted)

	if err != nil {
		t.Fatal(err)
	}

	var p payload

	if err := json.Unmarshal(decoded, &p); err != nil {
		t.Fatal(err)
	}

	return p
}

func encodePayload(t *testing.T, p payload) string {
	t.Helper()

	js, err := json.Marshal(p)

	if err != nil {
		t.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(js)
}
