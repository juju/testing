// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package filetesting

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/testing"
)

type StubFile struct {
	Stub *testing.Stub

	ReturnRead  io.Reader
	ReturnWrite io.Writer
	ReturnName  string
	ReturnSeek  int64
	ReturnStat  os.FileInfo
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

	return s.ReturnName
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

	return s.ReturnStat, nil
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
