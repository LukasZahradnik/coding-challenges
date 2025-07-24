package domain

type SignatureAlgorithm string

const (
	ECCAlgorithm SignatureAlgorithm = "ECC"
	RSAAlgorithm SignatureAlgorithm = "RSA"
)

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

type SignatureDevice struct {
	ID    string
	Label string

	Algorithm  SignatureAlgorithm
	PrivateKey []byte
	PublicKey  []byte

	Counter       int64
	LastSignature string
}
