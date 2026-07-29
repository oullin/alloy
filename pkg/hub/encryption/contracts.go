package encryption

import cencryption "hara.sh/alloy/contracts/encryption"

// EncrypterContract encrypts and decrypts values.
type EncrypterContract = cencryption.EncrypterContract

// StringEncrypterContract encrypts and decrypts strings without serialization.
type StringEncrypterContract = cencryption.StringEncrypterContract
