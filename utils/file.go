package utils

import (
	"airbox/config"
	"strconv"
	"strings"
)

// GetFileType return the file type according to the file suffix
func GetFileType(suffix string) config.FileType {
	if suffix == ".txt" || suffix == ".doc" || suffix == ".docx" || suffix == ".html" || suffix == ".pdf" ||
		suffix == ".xls" || suffix == ".xlsx" || suffix == ".ppt" || suffix == ".pptx" || suffix == ".md" {
		return config.FileDocumentType
	} else if ".bmp" == suffix || ".gif" == suffix || ".jpg" == suffix || ".png" == suffix || ".jepg" == suffix ||
		".svg" == suffix || ".tiff" == suffix || ".psd" == suffix || ".raw" == suffix || ".eps" == suffix {
		return config.FilePictureType
	} else if ".avi" == suffix || ".mov" == suffix || ".mkv" == suffix || ".asf" == suffix || ".rmvb" == suffix ||
		".mpeg" == suffix || ".wmv" == suffix || ".mp4" == suffix || ".ts" == suffix || ".flv" == suffix {
		return config.FileVideoType
	} else if ".mp3" == suffix || ".wma" == suffix || ".wav" == suffix || "aac" == suffix || ".flac" == suffix ||
		".ape" == suffix || ".aiff" == suffix || ".ogg" == suffix {
		return config.FileMusicType
	} else {
		return config.FileOtherType
	}
}

// AddIndexToFilename used to rename the file if the file is exist
func AddIndexToFilename(file string, index int) string {
	split := strings.LastIndex(file, ".")
	return file[:split] + "(" + strconv.FormatInt(int64(index), 10) + ")" + file[split:]
}
