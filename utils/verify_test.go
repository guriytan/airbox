package utils

import (
	"fmt"
	"testing"
)

func TestMd5Sum(t *testing.T) {
	md5 := "60e49a7763d879e8651a8836ae09f4aa"
	result, err := MD5Sum("C:/Users/guriytan/Downloads/Yes,Minister.S1-3/[是,.大臣].Yes,Minister.S01E01(ED2000.COM).rmvb")
	if err != nil {
		fmt.Printf("failed to get md5: %v\n", err)
		return
	}
	fmt.Printf("md5 fo file is: %v\n", result)
	if result != md5 {
		fmt.Printf("md5 is not equal\n")
	}
}

func TestSha256Sum(t *testing.T) {
	sha256 := "831d460c84d4c2ea28b5b826b87c0498493833a2b005a6e2e8be349eecc80872"
	result, err := SHA256Sum("C:/Users/guriytan/Downloads/Yes,Minister.S1-3/[是,.大臣].Yes,Minister.S01E01(ED2000.COM).rmvb")
	if err != nil {
		fmt.Printf("failed to get sha256: %v\n", err)
		return
	}
	fmt.Printf("sha256 fo file is: %v\n", result)
	if result != sha256 {
		fmt.Printf("sha256 is not equal\n")
	}
	fmt.Println(len(result))
}
