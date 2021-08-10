package mediatype

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

type SnifferContent interface {
	Read() []byte
	Stream() io.Reader
}

// Used to sniff a local file.
type SnifferFileContent struct {
	file *os.File
}

func NewSnifferFileContent(file *os.File) SnifferFileContent {
	return SnifferFileContent{file: file}
}

const MAX_READ_SIZE = 5 * 1024 * 1024 // 5MB

func (s SnifferFileContent) Read() []byte {
	s.file.Seek(0, io.SeekStart)
	info, err := s.file.Stat()
	if err != nil {
		return nil
	}
	if info.Size() > MAX_READ_SIZE {
		return nil
	}
	data := make([]byte, info.Size())
	_, err = s.file.Read(data)
	if err != nil && err != io.EOF {
		return nil
	}
	return data
}

func (s SnifferFileContent) Stream() io.Reader {
	s.file.Seek(0, io.SeekStart)
	return bufio.NewReader(s.file)
}

// Used to sniff a byte array.
type SnifferBytesContent struct {
	bytes []byte
}

func NewSnifferBytesContent(bytes []byte) SnifferBytesContent {
	return SnifferBytesContent{bytes: bytes}
}

func (s SnifferBytesContent) Read() []byte {
	return s.bytes
}

func (s SnifferBytesContent) Stream() io.Reader {
	return bytes.NewReader(s.bytes)
}

// TODO SnifferUriContent equivalent
