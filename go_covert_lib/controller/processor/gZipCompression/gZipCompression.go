package gZipCompression

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

type GZipCompression struct{}

//Compress a passed in byte array using GZip Compression
func (g *GZipCompression) Process(data []byte) ([]byte, error) {

	//buffer used to create a place to be written to
	var compData bytes.Buffer

	//Create a writer to compress
	gz := gzip.NewWriter(&compData)

	//Use the writer to compress the passed in data
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	//Close the writer
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return compData.Bytes(), nil
}

//Uncompress a passed in byte array using GZip Compression
func (g *GZipCompression) Unprocess(data []byte) ([]byte, error) {

	//Create a reader to uncompress
	gz, _ := gzip.NewReader(bytes.NewReader(data))

	//Use the reader to uncompress the passed in data
	unCompData, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	//Close the reader
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return unCompData, nil
}
