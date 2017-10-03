package file

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test seek position with read/write
func TestSeek(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_wrapper")
	checkErr(err)
	defer func() {
		err := os.RemoveAll(dir)
		checkErr(err)
	}()

	fn := filepath.Join(dir, "test.log")
	f, err := openFile(fn)
	checkErr(err)

	fw := NewFileWrapper(f)

	assert.Equal(t, int64(0), fw.pos)

	str := "This is a test.\n"
	n, err := fw.WriteString(str)
	checkErr(err)

	assert.Equal(t, len(str), n)
	assert.Equal(t, int64(n), fw.pos)

	str2 := "Foooooooooooooooooooooooooo\n"
	_, err = fw.WriteString(str2)
	checkErr(err)

	lines, err := readLines(fn)
	checkErr(err)
	assert.Equal(t, lines[0], str[0:len(str)-1])
	assert.Equal(t, lines[1], str2[0:len(str2)-1])
}

func TestReadAtLeast(t *testing.T) {
	dir, err := ioutil.TempDir("", "file_wrapper")
	checkErr(err)
	defer func() {
		err := os.RemoveAll(dir)
		checkErr(err)
	}()

	fn := filepath.Join(dir, "test.log")
	f, err := openFile(fn)
	checkErr(err)

	fw := NewFileWrapper(f)

	str := "0123456789\n"
	_, err = fw.WriteString(str)
	checkErr(err)

	fw.SeekAbs(0)
	var size int64 = 4
	n, b, err := fw.ReadAtLeast(size)
	assert.NoError(t, err)
	assert.Equal(t, size, n)
	assert.Equal(t, []byte("0123"), b)

	n, b, err = fw.ReadAtLeast(size)
	assert.NoError(t, err)
	assert.Equal(t, size, n)
	assert.Equal(t, []byte("4567"), b)

	n, b, err = fw.ReadAtLeast(size)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), n)
	assert.Equal(t, []byte("89\n"), b)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func openFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_SYNC|os.O_CREATE, 0644)
}

func readLines(path string) ([]string, error) {
	var lines []string

	f, err := os.Open(path)
	if err != nil {
		return lines, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, err
}
