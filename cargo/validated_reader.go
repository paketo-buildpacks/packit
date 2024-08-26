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
)

var ChecksumValidationError = errors.New("validation error: checksum does not match")

type ValidatedReader struct {
	reader   io.Reader
	checksum Checksum
	hash     hash.Hash
}

type errorHash struct {
	hash.Hash

	err error
}

func NewValidatedReader(reader io.Reader, sum string) ValidatedReader {
	var hash hash.Hash
	checksum := Checksum(sum)

	switch checksum.Algorithm() {
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	default:
		return ValidatedReader{hash: errorHash{err: fmt.Errorf("unsupported algorithm %q: the following algorithms are supported [sha256, sha512]", checksum.Algorithm())}}
	}

	return ValidatedReader{
		reader:   reader,
		checksum: checksum,
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
		if sum != vr.checksum.Hash() {
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
