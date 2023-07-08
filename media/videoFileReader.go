package media

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

/*
This file contains code originally written by FiskFan1999, licensed under the MIT License.
The original code can be found at: https://github.com/FiskFan1999/peertube-multipart-upload
*/
type VideoFileByteCounter int64

const (
	VideoChunkSize VideoFileByteCounter = 1024 * 1024 * 2
)

type VideoFileReader struct {
	VideoFile       *os.File
	TotalBytes      VideoFileByteCounter
	ChunkSize       VideoFileByteCounter
	CurrentMinBytes VideoFileByteCounter

	/*
		Note: CurrentMaxBytes is calculated from
		length of byte array
	*/
}

func GetVideoFileReader(filename string, chunkSize VideoFileByteCounter) (vfr *VideoFileReader, err error) {
	vfr = new(VideoFileReader)

	if vfr.VideoFile, err = os.Open(filename); err != nil {
		return
	}

	// _, _ = GetContentType(vfr.VideoFile)

	/*
		Calculate total bytes
	*/
	size, err := GetFileSize(filename)
	if err != nil {
		return
	}
	vfr.TotalBytes = VideoFileByteCounter(size)

	/*
		Set ChunkSize, and set CurrentMinBytes
	*/
	vfr.CurrentMinBytes = 0
	vfr.ChunkSize = chunkSize

	return
}

type VFRCurrentChunk struct {
	Bytes       []byte
	RangeHeader string
	MinByte     VideoFileByteCounter
	MaxByte     VideoFileByteCounter
	Length      int
	Finished    bool
}

func GetContentType(f *os.File) (contentType string, err error) {

	mime, err := mimetype.DetectReader(f)
	if err != nil {
		return contentType, err
	}
	if mime.Parent() != nil {
		contentType = mime.Parent().String()
		fmt.Println(mime.Parent().String())
	} else {
		contentType = mime.String()
	}

	return contentType, nil

}
func GetFileSize(filename string) (int64, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return int64(-1), err
	}
	return file.Size(), nil // size in bytes
}

func (vfr *VideoFileReader) GetNextChunk() (res *VFRCurrentChunk, err error) {
	res = new(VFRCurrentChunk)
	res.MinByte = vfr.CurrentMinBytes
	res.Bytes = make([]byte, vfr.ChunkSize)
	numBytes, err := vfr.VideoFile.Read(res.Bytes)
	if numBytes != len(res.Bytes) {
		newB := make([]byte, numBytes)
		for i := range newB {
			newB[i] = res.Bytes[i]
		}
		res.Bytes = newB
	}
	if errors.Is(err, io.EOF) {
		/*
			Previous iteration read all the way to EOF.
		*/
		res.Finished = true
		res.MaxByte = res.MinByte // convert int numBytes to int64
		return res, nil
	} else if err != nil {
		return
	} else {
		res.MaxByte = res.MinByte + VideoFileByteCounter(numBytes) // convert int numBytes to int64
	}

	res.Length = numBytes
	res.MaxByte = res.MinByte + VideoFileByteCounter(numBytes) - 1 // convert int numBytes to int64
	res.RangeHeader = fmt.Sprintf("bytes %d-%d/%d", res.MinByte, res.MaxByte, vfr.TotalBytes)
	res.Finished = false
	vfr.CurrentMinBytes += VideoFileByteCounter(numBytes)
	return
}
