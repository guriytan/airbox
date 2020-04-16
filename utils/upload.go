package utils

import (
	. "airbox/config"
	"errors"
	"io"
	"os"
)

// Chunk is the file chunk definition
type Chunk struct {
	Offset uint64 // Chunk offset
	Size   uint64 // Chunk size.
	Buffer []byte // Chunk data
}

type Chunks []Chunk

func NewChunks(offset, size uint64, buffer []byte) Chunk {
	chunk := Chunk{
		Offset: offset,
		Size:   size,
		Buffer: make([]byte, size),
	}
	copy(chunk.Buffer, buffer[:size])
	return chunk
}

// Upload need to know the path to store file, the data stream and the length of data.
// die chan is used to control the upload and the listener is used to notify the progress.
func Upload(filepath, filename string, src io.Reader, contentLength uint64) error {
	tempFilePath := filepath + filename + FileTempSuffix

	_ = os.MkdirAll(filepath, os.ModePerm)
	// If the file does not exist, create one. If exists, the download will overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, FilePermMode)
	if err != nil {
		_ = os.RemoveAll(filepath)
		return err
	}
	_ = fd.Close()

	jobs := make(chan Chunk, 2*FileGoroutine)
	results := make(chan Chunk, 2*FileGoroutine)
	failed := make(chan error)
	count := make(chan bool)
	stop := make(chan bool)

	// Start the upload workers
	for w := 1; w <= FileGoroutine; w++ {
		go UploadWorker(tempFilePath, jobs, results, failed, stop, count)
	}

	// Upload parts concurrently
	go UploadScheduler(src, jobs, failed)

	defer CleanTemp(tempFilePath, filepath, count)

	// Waiting for parts upload finished
	timestamp := Epoch()
	var completedBytes uint64
	for {
		select {
		case part := <-results:
			completedBytes += part.Size
			timestamp = Epoch()
		case err := <-failed:
			close(stop)
			return err
		default:
		}

		// Handle timeout
		if Epoch()-timestamp > FileTimeout {
			close(stop)
			return errors.New("timeout for transmission")
		}

		if completedBytes >= contentLength {
			break
		}
	}

	return os.Rename(tempFilePath, filepath+filename)
}

// CleanTemp clean the temp file after something causing the upload progress is stop
func CleanTemp(tempFilePath, filepath string, count chan bool) {
	for i := 0; i < FileGoroutine; i++ {
		<-count
	}
	// if something wrong, the temp file bill be removed.
	if _, err := os.Stat(tempFilePath); err != nil {
		if !os.IsNotExist(err) {
			_ = os.RemoveAll(filepath)
		}
	} else {
		_ = os.RemoveAll(filepath)
	}
}

// UploadWorker read the data from the chan and write to the file parallel
func UploadWorker(filePathTemp string, jobs <-chan Chunk, results chan<- Chunk, failed chan<- error,
	stop <-chan bool, count chan<- bool) {
	defer func() {
		count <- true
	}()
	for part := range jobs {
		select {
		case <-stop:
			return
		default:
		}

		fd, err := os.OpenFile(filePathTemp, os.O_WRONLY, FilePermMode)
		if err != nil {
			failed <- err
			break
		}

		_, err = fd.Seek(int64(part.Offset), io.SeekStart)
		if err != nil {
			_ = fd.Close()
			failed <- err
			break
		}

		_, err = fd.Write(part.Buffer[:part.Size])
		if err != nil {
			_ = fd.Close()
			failed <- err
			break
		}
		_ = fd.Close()
		part.Buffer = nil

		results <- part
	}
}

// UploadWorker produce the data remaining the worker writes to the file
func UploadScheduler(src io.Reader, jobs chan Chunk, failed chan error) {
	buffer := make([]byte, FileDownloadPartSize)
	offset := uint64(0)
	defer close(jobs)
	for {
		// generate buffer data by reading circularly
		size, err := src.Read(buffer)
		if err != nil {
			if err == io.EOF {
				jobs <- NewChunks(offset, uint64(size), buffer)
				return
			}
			failed <- err
			return
		}
		jobs <- NewChunks(offset, uint64(size), buffer)
		offset += uint64(size)
	}
}
