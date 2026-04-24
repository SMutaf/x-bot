package middleware

import (
	"log"
)

func RecoveryWrapper(taskName string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Panic oluştuğunda burası çalışır
			log.Printf("KRİTİK HATA [%s]: %v", taskName, r)
		}
	}()

	fn()
}
