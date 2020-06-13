package disk

import (
	"airbox/model"
	"archive/zip"
	"io"
	"os"
)

// Compress 压缩多文件
// files 文件数组，可以是不同dir下的文件或者文件夹
// dest 压缩文件存放地址
func Compress(files chan *model.File, dest io.Writer) error {
	w := zip.NewWriter(dest)
	defer func() {
		_ = w.Close()
	}()
	for file := range files {
		open, err := os.Open(file.FileInfo.Path + file.FileInfo.Name)
		if err != nil {
			return err
		}
		err = compress(open, "", w)
		_ = open.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = prefix + "/" + header.Name
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}
	return nil
}
