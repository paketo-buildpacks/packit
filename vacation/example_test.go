package vacation_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"path/filepath"

	dsnetBzip2 "github.com/dsnet/compress/bzip2"
	"github.com/paketo-buildpacks/packit/v2/vacation"
	"github.com/ulikunitz/xz"
)

type ArchiveFile struct {
	Name    string
	Content []byte
}

func ExampleArchive() {
	tarBuffer := bytes.NewBuffer(nil)
	tw := tar.NewWriter(tarBuffer)

	tarFiles := []ArchiveFile{
		{Name: "some-tar-dir/"},
		{Name: "some-tar-dir/some-tar-file", Content: []byte("some-tar-dir/some-tar-file")},
		{Name: "tar-file", Content: []byte("tar-file")},
	}

	for _, file := range tarFiles {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()

	zipBuffer := bytes.NewBuffer(nil)
	zw := zip.NewWriter(zipBuffer)

	zipFiles := []ArchiveFile{
		{Name: "some-zip-dir/"},
		{Name: "some-zip-dir/some-zip-file", Content: []byte("some-zip-dir/some-zip-file")},
		{Name: "zip-file", Content: []byte("zip-file")},
	}

	for _, file := range zipFiles {
		header := &zip.FileHeader{Name: file.Name}
		header.SetMode(0755)

		f, err := zw.CreateHeader(header)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write(file.Content); err != nil {
			log.Fatal(err)
		}
	}

	zw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewArchive(bytes.NewReader(tarBuffer.Bytes()))
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	archive = vacation.NewArchive(bytes.NewReader(zipBuffer.Bytes()))
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// some-tar-dir/some-tar-file
	// some-zip-dir/some-zip-file
	// tar-file
	// zip-file
}

func ExampleArchive_StripComponents() {
	tarBuffer := bytes.NewBuffer(nil)
	tw := tar.NewWriter(tarBuffer)

	tarFiles := []ArchiveFile{
		{Name: "some-tar-dir/"},
		{Name: "some-tar-dir/some-tar-file", Content: []byte("some-tar-dir/some-tar-file")},
		{Name: "tar-file", Content: []byte("tar-file")},
	}

	for _, file := range tarFiles {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()

	zipBuffer := bytes.NewBuffer(nil)
	zw := zip.NewWriter(zipBuffer)

	zipFiles := []ArchiveFile{
		{Name: "some-zip-dir/"},
		{Name: "some-zip-dir/some-zip-file", Content: []byte("some-zip-dir/some-zip-file")},
		{Name: "zip-file", Content: []byte("zip-file")},
	}

	for _, file := range zipFiles {
		header := &zip.FileHeader{Name: file.Name}
		header.SetMode(0755)

		f, err := zw.CreateHeader(header)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write(file.Content); err != nil {
			log.Fatal(err)
		}
	}

	zw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewArchive(bytes.NewReader(tarBuffer.Bytes())).StripComponents(1)
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	archive = vacation.NewArchive(bytes.NewReader(zipBuffer.Bytes())).StripComponents(1)
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// some-tar-file
	// some-zip-file
}

func ExampleTarArchive() {
	buffer := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buffer)

	files := []ArchiveFile{
		{Name: "some-dir/"},
		{Name: "some-dir/some-other-dir/"},
		{Name: "some-dir/some-other-dir/some-file", Content: []byte("some-dir/some-other-dir/some-file")},
		{Name: "first", Content: []byte("first")},
		{Name: "second", Content: []byte("second")},
		{Name: "third", Content: []byte("third")},
	}

	for _, file := range files {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewTarArchive(bytes.NewReader(buffer.Bytes()))
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// first
	// second
	// some-dir/some-other-dir/some-file
	// third
}

func ExampleTarArchive_StripComponents() {
	buffer := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buffer)

	files := []ArchiveFile{
		{Name: "some-dir/"},
		{Name: "some-dir/some-other-dir/"},
		{Name: "some-dir/some-other-dir/some-file", Content: []byte("some-dir/some-other-dir/some-file")},
		{Name: "first", Content: []byte("first")},
		{Name: "second", Content: []byte("second")},
		{Name: "third", Content: []byte("third")},
	}

	for _, file := range files {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewTarArchive(bytes.NewReader(buffer.Bytes())).StripComponents(1)
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// some-other-dir/some-file
}

func ExampleGzipArchive() {
	buffer := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(buffer)

	_, err := gw.Write([]byte(`Gzip file contents`))
	if err != nil {
		log.Fatal(err)
	}

	gw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewGzipArchive(bytes.NewReader(buffer.Bytes())).WithName("gzip-file")
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destination, "gzip-file"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(content))

	// Output:
	// Gzip file contents
}

func ExampleGzipArchive_StripComponents() {
	buffer := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(buffer)
	tw := tar.NewWriter(gw)

	files := []ArchiveFile{
		{Name: "some-dir/"},
		{Name: "some-dir/some-other-dir/"},
		{Name: "some-dir/some-other-dir/some-file", Content: []byte("some-dir/some-other-dir/some-file")},
		{Name: "first", Content: []byte("first")},
		{Name: "second", Content: []byte("second")},
		{Name: "third", Content: []byte("third")},
	}

	for _, file := range files {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()
	gw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewGzipArchive(bytes.NewReader(buffer.Bytes())).StripComponents(1)
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// some-other-dir/some-file
}

func ExampleXZArchive() {
	buffer := bytes.NewBuffer(nil)
	xw, err := xz.NewWriter(buffer)
	if err != nil {
		log.Fatal(err)
	}

	_, err = xw.Write([]byte(`XZ file contents`))
	if err != nil {
		log.Fatal(err)
	}

	xw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewXZArchive(bytes.NewReader(buffer.Bytes())).WithName("xz-file")
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	contents, err := os.ReadFile(filepath.Join(destination, "xz-file"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(contents))

	// Output:
	// XZ file contents
}

func ExampleXZArchive_StripComponents() {
	buffer := bytes.NewBuffer(nil)
	xw, err := xz.NewWriter(buffer)
	if err != nil {
		log.Fatal(err)
	}

	tw := tar.NewWriter(xw)

	files := []ArchiveFile{
		{Name: "some-dir/"},
		{Name: "some-dir/some-other-dir/"},
		{Name: "some-dir/some-other-dir/some-file", Content: []byte("some-dir/some-other-dir/some-file")},
		{Name: "first", Content: []byte("first")},
		{Name: "second", Content: []byte("second")},
		{Name: "third", Content: []byte("third")},
	}

	for _, file := range files {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()
	xw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewXZArchive(bytes.NewReader(buffer.Bytes())).StripComponents(1)
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// some-other-dir/some-file
}

func ExampleBzip2Archive() {
	buffer := bytes.NewBuffer(nil)

	// Using the dsnet library because the Go compression library does not
	// have a writer. There is recent discussion on this issue
	// https://github.com/golang/go/issues/4828 to add an encoder. The
	// library should be removed once there is a native encoder
	bz, err := dsnetBzip2.NewWriter(buffer, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bz.Write([]byte(`Bzip2 file contents`))
	if err != nil {
		log.Fatal(err)
	}

	bz.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewBzip2Archive(bytes.NewReader(buffer.Bytes())).WithName("bzip2-file")
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	contents, err := os.ReadFile(filepath.Join(destination, "bzip2-file"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(contents))

	// Output:
	// Bzip2 file contents
}

func ExampleBzip2Archive_StripComponents() {
	buffer := bytes.NewBuffer(nil)

	// Using the dsnet library because the Go compression library does not
	// have a writer. There is recent discussion on this issue
	// https://github.com/golang/go/issues/4828 to add an encoder. The
	// library should be removed once there is a native encoder
	bz, err := dsnetBzip2.NewWriter(buffer, nil)
	if err != nil {
		log.Fatal(err)
	}

	tw := tar.NewWriter(bz)

	files := []ArchiveFile{
		{Name: "some-dir/"},
		{Name: "some-dir/some-other-dir/"},
		{Name: "some-dir/some-other-dir/some-file", Content: []byte("some-dir/some-other-dir/some-file")},
		{Name: "first", Content: []byte("first")},
		{Name: "second", Content: []byte("second")},
		{Name: "third", Content: []byte("third")},
	}

	for _, file := range files {
		err := tw.WriteHeader(&tar.Header{Name: file.Name, Mode: 0755, Size: int64(len(file.Content))})
		if err != nil {
			log.Fatal(err)
		}

		_, err = tw.Write(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw.Close()
	bz.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewBzip2Archive(bytes.NewReader(buffer.Bytes())).StripComponents(1)
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// some-other-dir/some-file
}

func ExampleZipArchive() {
	buffer := bytes.NewBuffer(nil)
	zw := zip.NewWriter(buffer)

	files := []ArchiveFile{
		{Name: "some-dir/"},
		{Name: "some-dir/some-other-dir/"},
		{Name: "some-dir/some-other-dir/some-file", Content: []byte("some-dir/some-other-dir/some-file")},
		{Name: "first", Content: []byte("first")},
		{Name: "second", Content: []byte("second")},
		{Name: "third", Content: []byte("third")},
	}

	for _, file := range files {
		header := &zip.FileHeader{Name: file.Name}
		header.SetMode(0755)

		f, err := zw.CreateHeader(header)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write(file.Content); err != nil {
			log.Fatal(err)
		}
	}

	zw.Close()

	destination, err := os.MkdirTemp("", "destination")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(destination)

	archive := vacation.NewZipArchive(bytes.NewReader(buffer.Bytes()))
	if err := archive.Decompress(destination); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			rel, err := filepath.Rel(destination, path)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s\n", rel)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// first
	// second
	// some-dir/some-other-dir/some-file
	// third
}
