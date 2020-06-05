package utils

import (
	"airbox/global"
)

// GetFileType return the file type according to the file suffix
func GetFileType(suffix string) global.FileType {
	if suffix == ".txt" || suffix == ".doc" || suffix == ".docx" || suffix == ".html" || suffix == ".pdf" ||
		suffix == ".xls" || suffix == ".xlsx" || suffix == ".ppt" || suffix == ".pptx" || suffix == ".md" {
		return global.FileDocumentType
	} else if ".bmp" == suffix || ".gif" == suffix || ".jpg" == suffix || ".png" == suffix || ".jepg" == suffix ||
		".svg" == suffix || ".tiff" == suffix || ".psd" == suffix || ".raw" == suffix || ".eps" == suffix {
		return global.FilePictureType
	} else if ".avi" == suffix || ".mov" == suffix || ".mkv" == suffix || ".asf" == suffix || ".rmvb" == suffix ||
		".mpeg" == suffix || ".wmv" == suffix || ".mp4" == suffix || ".ts" == suffix || ".flv" == suffix {
		return global.FileVideoType
	} else if ".mp3" == suffix || ".wma" == suffix || ".wav" == suffix || "aac" == suffix || ".flac" == suffix ||
		".ape" == suffix || ".aiff" == suffix || ".ogg" == suffix {
		return global.FileMusicType
	} else {
		return global.FileOtherType
	}
}
