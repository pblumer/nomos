package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/nomos/nomos/internal/storage"
	"gopkg.in/yaml.v3"
)

type apiKeyFile struct {
	Keys []apiKeyEntry `yaml:"keys"`
}

type apiKeyEntry struct {
	Name    string    `yaml:"name"`
	Hash    string    `yaml:"hash"`
	Created time.Time `yaml:"created"`
}

type APIKeyDTO struct {
	Name    string    `yaml:"name" json:"name"`
	Created time.Time `yaml:"created" json:"created"`
}

type CreateKeyResult struct {
	Key string `json:"key"`
	APIKeyDTO
}

func CreateKey(cosmosPath, name string) (CreateKeyResult, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return CreateKeyResult{}, err
	}
	key := "nomos_" + hex.EncodeToString(raw)
	h := sha256.Sum256([]byte(key))
	entry := apiKeyEntry{Name: name, Hash: hex.EncodeToString(h[:]), Created: time.Now().UTC()}

	kf, err := loadKeys(cosmosPath)
	if err != nil {
		return CreateKeyResult{}, err
	}
	for _, k := range kf.Keys {
		if k.Name == name {
			return CreateKeyResult{}, fmt.Errorf("key with name %q already exists", name)
		}
	}
	kf.Keys = append(kf.Keys, entry)
	if err := saveKeys(cosmosPath, kf); err != nil {
		return CreateKeyResult{}, err
	}
	return CreateKeyResult{Key: key, APIKeyDTO: APIKeyDTO{Name: name, Created: entry.Created}}, nil
}

func ListKeys(cosmosPath string) ([]APIKeyDTO, error) {
	kf, err := loadKeys(cosmosPath)
	if err != nil {
		return nil, err
	}
	out := make([]APIKeyDTO, len(kf.Keys))
	for i, k := range kf.Keys {
		out[i] = APIKeyDTO{Name: k.Name, Created: k.Created}
	}
	return out, nil
}

func RevokeKey(cosmosPath, name string) error {
	kf, err := loadKeys(cosmosPath)
	if err != nil {
		return err
	}
	next := kf.Keys[:0]
	found := false
	for _, k := range kf.Keys {
		if k.Name == name {
			found = true
		} else {
			next = append(next, k)
		}
	}
	if !found {
		return fmt.Errorf("key %q not found", name)
	}
	kf.Keys = next
	return saveKeys(cosmosPath, kf)
}

func ValidateKey(cosmosPath, key string) bool {
	kf, err := loadKeys(cosmosPath)
	if err != nil {
		return false
	}
	h := sha256.Sum256([]byte(key))
	hash := hex.EncodeToString(h[:])
	for _, k := range kf.Keys {
		if k.Hash == hash {
			return true
		}
	}
	return false
}

func loadKeys(cosmosPath string) (apiKeyFile, error) {
	path := storage.KeysFile(cosmosPath)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return apiKeyFile{}, nil
	}
	if err != nil {
		return apiKeyFile{}, err
	}
	var kf apiKeyFile
	if err := yaml.Unmarshal(data, &kf); err != nil {
		return apiKeyFile{}, err
	}
	return kf, nil
}

func saveKeys(cosmosPath string, kf apiKeyFile) error {
	path := storage.KeysFile(cosmosPath)
	data, err := yaml.Marshal(kf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
