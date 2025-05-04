package main

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base = int64(len(alphabet))

func EncodeBase62(num int64) string {
	if num == 0 {
		return string(alphabet[0])
	}
	result := ""
	for num > 0 {
		remainder := num % base
		result = string(alphabet[remainder]) + result
		num /= base
	}
	return result
}
