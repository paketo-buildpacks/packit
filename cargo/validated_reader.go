package cargo

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"io/ioutil"
)

var ChecksumValidationError = errors.New("validation error: checksum does not match")

type ValidatedReader struct {
	reader   io.Reader
	checksum string
	hash     hash.Hash
}

func NewValidatedReader(reader io.Reader, checksum string) ValidatedReader {
	return ValidatedReader{
		reader:   reader,
		checksum: checksum,
		hash:     sha256.New(),
	}
}

func (vr ValidatedReader) Read(p []byte) (int, error) {
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
	_, err := io.Copy(ioutil.Discard, vr)
	if err != nil {
		if err == ChecksumValidationError {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
