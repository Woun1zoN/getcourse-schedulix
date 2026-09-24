package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/Woun1zoN/getcourse-schedulix/internal/converter"
	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram"
)

func InitApp(chatID int64, client *getcourse.Client, state *storage.RedisState, telegramClient *telegram.Client) error {
	if err := client.ValidateCookie(); err != nil {
		return fmt.Errorf("validate cookie: %w", err)
	}

	lessonIDs, err := client.GetLessonIDs()
	if err != nil {
		return err
	}

	var checked atomic.Int64
	var newDocuments atomic.Int64
	var sent atomic.Int64

	fmt.Printf("Lessons: %d\n", len(lessonIDs))

	var tgMu sync.Mutex

	logs := make([]string, len(lessonIDs))

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(4)

	for i, id := range lessonIDs {
		i, id := i, id

		g.Go(func() error {
			var logBuf strings.Builder

			printf := func(format string, args ...any) {
				fmt.Fprintf(&logBuf, format, args...)
			}

			lessonURL := "https://shtpt.getcourse.ru/pl/teach/control/lesson/view?id=" +
				id +
				"&editMode=0"

			documents, err := client.GetLessonDocuments(id)
			if err != nil {
				printf("  Lesson ID: %s\n", id)
				printf("  ERROR: %v\n\n", err)
				logs[i] = logBuf.String()
				return nil
			}

			for _, document := range documents {
				checked.Add(1)

				name := strings.TrimPrefix(document.Name, "Занятия на ")
				name = strings.TrimSuffix(name, ".doc")

				data, err := client.DownloadDocument(document)
				if err != nil {
					printf("  Lesson ID: %s\n", document.ID)
					printf("  Document: %s\n", name)
					printf("  URL: %s\n", document.URL)
					printf("  ERROR Download: %v\n\n", err)
					continue
				}

				fileSize := float64(len(data)) / 1024

				changed, err := state.HasChanged(ctx, document.URL, data)
				if err != nil {
					printf("  Lesson ID: %s\n", document.ID)
					printf("  Document: %s\n", name)
					printf("  ERROR Redis HasChanged: %v\n\n", err)
					continue
				}

				if !changed {
					printf("  Lesson ID: %s\n", document.ID)
					printf("  Document: %s\n", name)
					printf("  URL: %s\n", document.URL)
					printf("  Downloaded: %d bytes\n", len(data))
					printf("  Already processed\n\n")
					continue
				}

				newDocuments.Add(1)

				printf("DETECTED NEW DOCUMENT\n")
				printf("  Lesson ID: %s\n", document.ID)
				printf("  Document: %s\n", name)
				printf("  URL: %s\n", document.URL)
				printf("  Downloaded: %d bytes\n", len(data))

				jpgPath, err := converter.ConvertToJPG(data, document.Name)
				if err != nil {
					printf("  ERROR Convert: %v\n\n", err)
					continue
				}

				printf("  JPG: %s\n", jpgPath)

				captionName := strings.TrimSuffix(
					name,
					".doc",
				)

				caption := fmt.Sprintf(
					"Расписание на %s\n\n"+
						"🔗 %.1f КБ | [Скачать](%s)",
					captionName,
					fileSize,
					document.URL,
				)

				tgMu.Lock()
				err = telegramClient.SendPhoto(chatID, jpgPath, caption)
				tgMu.Unlock()

				if err != nil {
					printf("  ERROR Telegram: %v\n\n", err)
					continue
				}

				sent.Add(1)

				if err := telegramClient.SendNewDocumentLog(document.Name, document.URL, document.ID, lessonURL, fileSize, checked.Load(), newDocuments.Load()); err != nil {
					log.Printf("telegram log %s: %v", document.Name, err)
				}

				if err := state.MarkProcessed(ctx, document.URL, data); err != nil {
					printf("  ERROR Redis MarkProcessed: %v\n\n", err)
				}

				printf("SENT TO TELEGRAM\n\n")
			}

			logs[i] = logBuf.String()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	for _, l := range logs {
		fmt.Print(l)
	}

	fmt.Println()
	fmt.Printf("Checked: %d\n", checked.Load())
	fmt.Printf("New: %d\n", newDocuments.Load())
	fmt.Printf("Sent: %d\n\n", sent.Load())

	return nil
}