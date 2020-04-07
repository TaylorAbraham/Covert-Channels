package zLibCompression

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

type ZLibCompression struct{}

//Compress a passed in byte array using ZLib Compression
func (z *ZLibCompression) Process(data []byte) ([]byte, error) {

	//buffer used to create a place to be written to
	var compData bytes.Buffer

	//Create a writer to compress
	zl := zlib.NewWriter(&compData)

	//Use the writer to compress the passed in data
	if _, err := zl.Write(data); err != nil {
		return nil, err
	}

	//Close the writer
	if err := zl.Close(); err != nil {
		return nil, err
	}

	return compData.Bytes(), nil
}

//Uncompress a passed in byte array using ZLib Compression
func (z *ZLibCompression) Unprocess(data []byte) ([]byte, error) {

	//Create a reader to uncompress
	zl, _ := zlib.NewReader(bytes.NewReader(data))

	//Use the reader to uncompress the passed in data
	unCompData, err := ioutil.ReadAll(zl)
	if err != nil {
		return nil, err
	}

	//Close the reader
	if err := zl.Close(); err != nil {
		return nil, err
	}

	return unCompData, nil
}
