package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"ai_tg_writer/internal/infrastructure/whisper"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run test_api.go <путь_к_аудио_файлу>")
		os.Exit(1)
	}

	audioPath := os.Args[1]

	// Создаем обработчик Whisper
	whisperHandler := whisper.NewWhisperHandler()

	fmt.Printf("Отправляем файл на транскрипцию: %s\n", audioPath)

	// Отправляем файл на транскрипцию
	transcriptionResp, err := whisperHandler.TranscribeAudio(audioPath)
	if err != nil {
		log.Fatalf("Ошибка отправки на транскрипцию: %v", err)
	}

	fmt.Printf("Файл отправлен на транскрипцию:\n")
	fmt.Printf("  File ID: %s\n", transcriptionResp.FileID)
	fmt.Printf("  Статус: %s\n", transcriptionResp.Status)
	fmt.Printf("  Позиция в очереди: %d\n", transcriptionResp.QueuePosition)
	fmt.Printf("  Размер очереди: %d\n", transcriptionResp.Metrics.QueueSize)
	fmt.Printf("  Среднее время обработки: %.2f сек\n", transcriptionResp.Metrics.AvgProcessingTime)

	// Ждем завершения транскрипции
	fmt.Println("\nОжидаем завершения транскрипции...")
	transcribedText, err := whisperHandler.WaitForCompletion(transcriptionResp.FileID, 5*time.Minute)
	if err != nil {
		log.Fatalf("Ошибка ожидания транскрипции: %v", err)
	}

	fmt.Printf("\n✅ Транскрипция завершена!\n")
	fmt.Printf("📝 Результат:\n%s\n", transcribedText)
}
