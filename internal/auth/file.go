package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"

	"github.com/jonasandre/movidesk-cli/internal/config"
)

const (
	// EnvPassphrase is read by the file fallback; if empty we use a fixed
	// machine-local key derived from the user home. The file is still chmod 0600.
	EnvPassphrase = "MOVIDESK_PASSPHRASE"

	credsFile  = "credentials.enc"
	pbkdfIters = 100_000
	saltLen    = 16
	nonceLen   = 12
	keyLen     = 32
)

// fileStore is an AES-GCM encrypted JSON map of tenant -> token, persisted at
// ~/.movidesk/credentials.enc (chmod 0600).
type fileStore struct{}

func newFileStore() Store { return &fileStore{} }

type credsBlob struct {
	Salt  []byte            `json:"salt"`
	Nonce []byte            `json:"nonce"`
	Data  []byte            `json:"data"` // ciphertext
	V     int               `json:"v"`
	M     map[string]string `json:"-"` // not serialized; in-memory cache
}

func credsPath() (string, error) {
	d, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, credsFile), nil
}

func passphrase() []byte {
	if v := os.Getenv(EnvPassphrase); v != "" {
		return []byte(v)
	}
	// Machine-local fallback: deterministic per user, but file is chmod 0600
	// and stays on disk. This is a soft fallback for headless Linux.
	//
	// Security note: the derived key is predictable — anyone with access to
	// the user home path can reproduce it. Set MOVIDESK_PASSPHRASE for a
	// stronger guarantee in headless or shared environments.
	fmt.Fprintf(os.Stderr,
		"aviso: %s não está definida; usando uma chave previsível e local pra criptografar credenciais. "+
			"Defina %s pra maior segurança em ambientes headless.\n",
		EnvPassphrase, EnvPassphrase)
	home, _ := os.UserHomeDir()
	h := sha256.Sum256([]byte("movidesk-cli|" + home))
	return h[:]
}

func deriveKey(pass, salt []byte) []byte {
	return pbkdf2.Key(pass, salt, pbkdfIters, keyLen, sha256.New)
}

func loadBlob() (map[string]string, *credsBlob, error) {
	p, err := credsPath()
	if err != nil {
		return nil, nil, err
	}
	raw, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, &credsBlob{V: 1}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("ler arquivo de credenciais: %w", err)
	}
	var blob credsBlob
	if err := json.Unmarshal(raw, &blob); err != nil {
		return nil, nil, fmt.Errorf("interpretar arquivo de credenciais: %w", err)
	}
	key := deriveKey(passphrase(), blob.Salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	plain, err := gcm.Open(nil, blob.Nonce, blob.Data, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("descriptografar credenciais: %w (tente definir %s)", err, EnvPassphrase)
	}
	m := map[string]string{}
	if err := json.Unmarshal(plain, &m); err != nil {
		return nil, nil, fmt.Errorf("decodificar credenciais: %w", err)
	}
	return m, &blob, nil
}

func saveBlob(m map[string]string) error {
	plain, err := json.Marshal(m)
	if err != nil {
		return err
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	key := deriveKey(passphrase(), salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	cipherText := gcm.Seal(nil, nonce, plain, nil)

	blob := credsBlob{V: 1, Salt: salt, Nonce: nonce, Data: cipherText}
	raw, err := json.Marshal(blob)
	if err != nil {
		return err
	}

	d, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	p := filepath.Join(d, credsFile)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (s *fileStore) Get(tenant string) (string, error) {
	m, _, err := loadBlob()
	if err != nil {
		return "", err
	}
	v, ok := m[tenant]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (s *fileStore) Set(tenant, token string) error {
	m, _, err := loadBlob()
	if err != nil {
		return err
	}
	m[tenant] = token
	return saveBlob(m)
}

func (s *fileStore) Delete(tenant string) error {
	m, _, err := loadBlob()
	if err != nil {
		return err
	}
	if _, ok := m[tenant]; !ok {
		return ErrNotFound
	}
	delete(m, tenant)
	if len(m) == 0 {
		p, err := credsPath()
		if err != nil {
			return err
		}
		return os.Remove(p)
	}
	return saveBlob(m)
}

// EncodePeek returns a redacted preview of a token for UI display.
func EncodePeek(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	sum := sha256.Sum256([]byte(token))
	return token[:4] + "…" + base64.RawURLEncoding.EncodeToString(sum[:3])
}
