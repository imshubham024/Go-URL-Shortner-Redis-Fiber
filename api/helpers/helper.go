package helpers

import(
	"os"
	"strings"
)

func EnforceHTTP(url string) string{
	if url[:4]!="http"{
		url="http://"+url
	}
	return url
}

func RemoveDomainError(url string) string{
	
}