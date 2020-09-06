// Copyright 2020 Frederik Zipp. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ggpack

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPackRoundTrip(t *testing.T) {
	filePackRoundTrip(t, "test.txt")
}

func TestPackRoundTripBnutFile(t *testing.T) {
	filePackRoundTrip(t, "test.bnut")
}

func filePackRoundTrip(t *testing.T, dataFileName string) {
	packFilePath := "test.ggpack"
	packFile, err := ioutil.TempFile("", packFilePath)
	if err != nil {
		t.Errorf("could not create pack file: %s", err)
	}
	defer packFile.Close()
	defer os.Remove(packFile.Name())

	packer, err := NewPacker(packFile)
	if err != nil {
		t.Errorf("could not create packer: %s", err)
	}

	dataFilePath := filepath.Join("testdata", dataFileName)
	original, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		t.Errorf("could not read test data from file: %s", err)
	}

	err = packer.WriteFile(dataFilePath)
	if err != nil {
		t.Errorf("could not write file to ggpack: %s", err)
	}
	err = packer.Finish()
	if err != nil {
		t.Errorf("could not finish ggpack: %s", err)
	}

	pack, err := Open(packFile.Name())
	if err != nil {
		t.Errorf("could not open ggpack: %s", err)
	}
	defer pack.Close()

	r, _, err := pack.File(dataFileName)
	if err != nil {
		t.Errorf("could not access file from ggpack: %s", err)
	}
	decoded, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("could not add file from ggpack: %s", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("decoded data is not equal to original data! Original: %q vs. decoded: %q", string(original), string(decoded))
	}
}
