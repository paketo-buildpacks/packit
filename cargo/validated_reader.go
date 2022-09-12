package cargo

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

var ChecksumValidationError = errors.New("validation error: checksum does not match")

type ValidatedReader struct {
	reader   io.Reader
	checksum string
	hash     hash.Hash
}

type errorHash struct {
	hash.Hash

	err error
}

func NewValidatedReader(reader io.Reader, checksum string) ValidatedReader {
	splitChecksum := strings.SplitN(checksum, ":", 2)
	if len(splitChecksum) != 2 {
		return ValidatedReader{hash: errorHash{err: fmt.Errorf(`malformed checksum %q: checksum should be formatted "algorithm:hash"`, checksum)}}
	}

	checksumValue := splitChecksum[1]

	var hash hash.Hash

	switch splitChecksum[0] {
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	default:
		return ValidatedReader{hash: errorHash{err: fmt.Errorf("unsupported algorithm %q: the following algorithms are support [sha256, sha512]", splitChecksum[0])}}
	}

	return ValidatedReader{
		reader:   reader,
		checksum: checksumValue,
		hash:     hash,
	}
}

func (vr ValidatedReader) Read(p []byte) (int, error) {
	if errHash, ok := vr.hash.(errorHash); ok {
		return 0, errHash.err
	}

	var done bool
	n, err := vr.reader.Read(p)
	if err != nil {
		if err == io.EOF {
			done = true
		} else {
			return n, err
		}
	}

	buffer := bytes.NewBuffer(p)
	_, err = io.CopyN(vr.hash, buffer, int64(n))
	if err != nil {
		return n, err
	}

	if done {
		sum := hex.EncodeToString(vr.hash.Sum(nil))
		if sum != vr.checksum {
			return n, ChecksumValidationError
		}

		return n, io.EOF
	}

	return n, nil
}

func (vr ValidatedReader) Valid() (bool, error) {
	_, err := io.Copy(io.Discard, vr)
	if err != nil {
		if err == ChecksumValidationError {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
