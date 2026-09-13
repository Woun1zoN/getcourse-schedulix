package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"sync"
	"sync/atomic"
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/joho/godotenv"

	"github.com/Woun1zoN/schedule-bot/internal/converter"
	"github.com/Woun1zoN/schedule-bot/internal/getcourse"
	"github.com/Woun1zoN/schedule-bot/internal/storage"
	"github.com/Woun1zoN/schedule-bot/internal/telegram"
)

func main() {
	_ = godotenv.Load()

	cookies := os.Getenv("GETCOURSE_COOKIES")

	client := getcourse.NewClient(
		"https://shtpt.getcourse.ru",
		cookies,
	)

	state, err := storage.Load("state.json")
	if err != nil {
		log.Fatal(err)
	}

	telegramClient := telegram.NewClient(
		os.Getenv("TELEGRAM_BOT_TOKEN"),
		os.Getenv("TELEGRAM_CHAT_ID"),
	)

	if err := runOnce(client, state, telegramClient); err != nil {
		log.Println("run:", err)
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := runOnce(client, state, telegramClient); err != nil {
			log.Println("run:", err)
		}
	}
}

func runOnce(client *getcourse.Client, state *storage.State, telegramClient *telegram.Client) error {
	lessonIDs, err := client.GetLessonIDs()
	if err != nil {
		return err
	}

	var checked atomic.Int64
	var newDocuments atomic.Int64
	var sent atomic.Int64

	fmt.Printf("\nLessons: %d\n", len(lessonIDs))

	var stateMu sync.Mutex
	var tgMu sync.Mutex

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(4)

	for _, id := range lessonIDs {
		id := id

		g.Go(func() error {
			fmt.Println("Lesson:", id)

			documents, err := client.GetLessonDocuments(id)
			if err != nil {
				log.Printf("lesson %s: %v", id, err)
				return nil
			}

			for _, document := range documents {
				checked.Add(1)

				fmt.Printf("  Document: %s\n", document.Name)
				fmt.Printf("  URL: %s\n", document.URL)

				data, err := client.DownloadDocument(document)
				if err != nil {
					log.Printf("download %s: %v", document.Name, err)
					continue
				}

				fmt.Printf("  Downloaded: %d bytes\n", len(data))

				stateMu.Lock()
				changed := state.HasChanged(document.URL, data)
				stateMu.Unlock()
				if !changed {
					fmt.Println("  Already processed")
					continue
				}

				newDocuments.Add(1)

				fmt.Println("  NEW DOCUMENT")

				jpgPath, err := converter.ConvertToJPG(data, document.Name)
				if err != nil {
					log.Printf("convert %s: %v", document.Name, err)
					continue
				}

				fmt.Println("  JPG:", jpgPath)

				name := strings.TrimSuffix(document.Name, ".doc")
				name = strings.TrimPrefix(name, "Занятия на ")

				caption := "📚 Расписание на " + name

				tgMu.Lock()
				err = telegramClient.SendPhoto(jpgPath, caption)
				tgMu.Unlock()
				if err != nil {
					log.Printf("telegram %s: %v", document.Name, err)
					continue
				}

				sent.Add(1)

				stateMu.Lock()
				state.MarkProcessed(document.URL, data)
				stateMu.Unlock()

				fmt.Println("  SENT TO TELEGRAM")
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
        return err
    }

	fmt.Println()
	fmt.Printf("Checked: %d\n", checked.Load())
	fmt.Printf("New: %d\n", newDocuments.Load())
	fmt.Printf("Sent: %d\n", sent.Load())

	if err := state.Save("state.json"); err != nil {
		return err
	}

	return nil
}