package vacation_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/ulikunitz/xz"
)

func ExampleTarArchive() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a tar byte stream on buffer.
	tw := tar.NewWriter(buffer)

	tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
	tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})
	tw.Write([]byte(nestedFile))

	for _, file := range []string{"first", "second", "third"} {
		tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})
		tw.Write([]byte(file))
	}

	tw.Close()

	//Running decompression
	vacation.NewTarArchive(bytes.NewReader(buffer.Bytes())).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/first
	// destination/second
	// destination/some-dir/some-other-dir/some-file
	// destination/third
}

func ExampleTarArchive_StripComponents() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a tar byte stream on buffer.
	tw := tar.NewWriter(buffer)

	tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
	tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})
	tw.Write([]byte(nestedFile))

	for _, file := range []string{"first", "second", "third"} {
		tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})
		tw.Write([]byte(file))
	}

	tw.Close()

	//Running decompression
	vacation.NewTarArchive(bytes.NewReader(buffer.Bytes())).StripComponents(1).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/some-other-dir/some-file
}

func ExampleTarGzipArchive() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a gzip tar byte stream on buffer.
	gw := gzip.NewWriter(buffer)
	tw := tar.NewWriter(gw)

	tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
	tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})
	tw.Write([]byte(nestedFile))

	for _, file := range []string{"first", "second", "third"} {
		tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})
		tw.Write([]byte(file))
	}

	tw.Close()
	gw.Close()

	//Running decompression
	vacation.NewTarGzipArchive(bytes.NewReader(buffer.Bytes())).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/first
	// destination/second
	// destination/some-dir/some-other-dir/some-file
	// destination/third
}

func ExampleTarGzipArchive_StripComponents() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a gzip tar byte stream on buffer.
	gw := gzip.NewWriter(buffer)
	tw := tar.NewWriter(gw)

	tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
	tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})
	tw.Write([]byte(nestedFile))

	for _, file := range []string{"first", "second", "third"} {
		tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})
		tw.Write([]byte(file))
	}

	tw.Close()
	gw.Close()

	//Running decompression
	vacation.NewTarGzipArchive(bytes.NewReader(buffer.Bytes())).StripComponents(1).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/some-other-dir/some-file
}

func ExampleTarXZArchive() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a xz tar byte stream on buffer.
	xw, _ := xz.NewWriter(buffer)
	tw := tar.NewWriter(xw)

	tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
	tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})
	tw.Write([]byte(nestedFile))

	for _, file := range []string{"first", "second", "third"} {
		tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})
		tw.Write([]byte(file))
	}

	tw.Close()
	xw.Close()

	//Running decompression
	vacation.NewTarXZArchive(bytes.NewReader(buffer.Bytes())).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/first
	// destination/second
	// destination/some-dir/some-other-dir/some-file
	// destination/third
}

func ExampleTarXZArchive_StripComponents() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a gzip tar byte stream on buffer.
	xw, _ := xz.NewWriter(buffer)
	tw := tar.NewWriter(xw)

	tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	tw.WriteHeader(&tar.Header{Name: filepath.Join("some-dir", "some-other-dir"), Mode: 0755, Typeflag: tar.TypeDir})
	tw.Write(nil)

	nestedFile := filepath.Join("some-dir", "some-other-dir", "some-file")
	tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})
	tw.Write([]byte(nestedFile))

	for _, file := range []string{"first", "second", "third"} {
		tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})
		tw.Write([]byte(file))
	}

	tw.Close()
	xw.Close()

	//Running decompression
	vacation.NewTarXZArchive(bytes.NewReader(buffer.Bytes())).StripComponents(1).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/some-other-dir/some-file
}

func ExampleZipArchive() {
	os.Mkdir("destination", os.ModePerm)
	defer os.RemoveAll("destination")

	buffer := bytes.NewBuffer(nil)

	// Constructing a zip byte stream on buffer.
	zw := zip.NewWriter(buffer)

	zw.Create("some-dir/")

	zw.Create(fmt.Sprintf("%s/", filepath.Join("some-dir", "some-other-dir")))

	fileHeader := &zip.FileHeader{Name: filepath.Join("some-dir", "some-other-dir", "some-file")}
	fileHeader.SetMode(0644)

	nestedFile, _ := zw.CreateHeader(fileHeader)
	nestedFile.Write([]byte("nested file"))

	for _, name := range []string{"first", "second", "third"} {
		fileHeader := &zip.FileHeader{Name: name}
		fileHeader.SetMode(0755)

		f, _ := zw.CreateHeader(fileHeader)
		f.Write([]byte(name))
	}

	zw.Close()

	//Running decompression
	vacation.NewZipArchive(bytes.NewReader(buffer.Bytes())).Decompress("destination")

	// Showing files in destination
	var files []string
	filepath.Walk("destination", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return nil
	})

	for _, f := range files {
		fmt.Printf("%s\n", f)
	}

	// Output:
	// destination/first
	// destination/second
	// destination/some-dir/some-other-dir/some-file
	// destination/third
}
