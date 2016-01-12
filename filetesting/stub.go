// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filetesting

import (
	"bytes"
	"io"
	"os"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/juju/testing"
)

type StubFile struct {
	Stub *testing.Stub

	Info        StubFileInfo
	ReturnRead  io.Reader
	ReturnWrite io.Writer
	ReturnSeek  int64
}

func NewStubReader(stub *testing.Stub, content string) io.Reader {
	return &StubFile{
		Stub:       stub,
		ReturnRead: strings.NewReader(content),
	}
}

func NewStubWriter(stub *testing.Stub) (io.Writer, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	s := &StubFile{
		Stub:        stub,
		ReturnWrite: buf,
	}
	return s, buf
}

func (s *StubFile) Read(data []byte) (int, error) {
	s.Stub.AddCall("Read", data)
	if err := s.Stub.NextErr(); err != nil {
		return 0, errors.Trace(err)
	}

	return s.ReturnRead.Read(data)
}

func (s *StubFile) Write(data []byte) (int, error) {
	s.Stub.AddCall("Write", data)
	if err := s.Stub.NextErr(); err != nil {
		return 0, errors.Trace(err)
	}

	return s.ReturnWrite.Write(data)
}

func (s *StubFile) Name() string {
	s.Stub.AddCall("Name")
	s.Stub.NextErr() // Pop one off.

	return s.Info.Info.Name
}

func (s *StubFile) Seek(offset int64, whence int) (int64, error) {
	s.Stub.AddCall("Seek", offset, whence)
	if err := s.Stub.NextErr(); err != nil {
		return 0, errors.Trace(err)
	}

	return s.ReturnSeek, nil
}

func (s *StubFile) Stat() (os.FileInfo, error) {
	s.Stub.AddCall("Stat")
	if err := s.Stub.NextErr(); err != nil {
		return nil, errors.Trace(err)
	}

	return &s.Info, nil
}

func (s *StubFile) Sync() error {
	s.Stub.AddCall("Sync")
	if err := s.Stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *StubFile) Truncate(size int64) error {
	s.Stub.AddCall("Truncate", size)
	if err := s.Stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *StubFile) Close() error {
	s.Stub.AddCall("Close")
	if err := s.Stub.NextErr(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
}

var _ os.FileInfo = (*StubFileInfo)(nil)

type StubFileInfo struct {
	Stub *testing.Stub

	Info      FileInfo
	ReturnSys interface{}
}

func NewStubFileInfo(stub *testing.Stub, name, content string) *StubFileInfo {
	return &StubFileInfo{
		Stub: stub,
		Info: FileInfo{
			Name:    name,
			Size:    int64(len(content)),
			Mode:    0644,
			ModTime: time.Now(),
		},
	}
}

func (s StubFileInfo) Name() string {
	s.Stub.AddCall("Name")
	s.Stub.NextErr() // Pop one off.

	return s.Info.Name
}

func (s StubFileInfo) Size() int64 {
	s.Stub.AddCall("Size")
	s.Stub.NextErr() // Pop one off.

	return s.Info.Size
}

func (s StubFileInfo) Mode() os.FileMode {
	s.Stub.AddCall("Mode")
	s.Stub.NextErr() // Pop one off.

	return s.Info.Mode
}

func (s StubFileInfo) ModTime() time.Time {
	s.Stub.AddCall("ModTime")
	s.Stub.NextErr() // Pop one off.

	return s.Info.ModTime
}

func (s StubFileInfo) IsDir() bool {
	s.Stub.AddCall("IsDir")
	s.Stub.NextErr() // Pop one off.

	return s.Info.Mode.IsDir()
}

func (s StubFileInfo) Sys() interface{} {
	s.Stub.AddCall("Sys")
	s.Stub.NextErr() // Pop one off.

	return s.ReturnSys
}
