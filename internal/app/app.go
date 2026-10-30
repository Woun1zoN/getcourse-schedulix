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
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
)

func InitApp(ctx context.Context, chatID, targetChatID, streamID int64, client *getcourse.Client, state *storage.RedisState, telegramClient *botapi.Client, notifyMaterial func(documentName, documentURL, lessonID, lessonURL string, fileSize float64, checked, newDocuments, logChatID int64) error) error {
	if err := client.ValidateCookie(ctx); err != nil {
		return fmt.Errorf("validate cookie: %w", err)
	}

	lessonIDs, err := client.GetLessonIDs(ctx, streamID)
	if err != nil {
		return err
	}

	baselined, err := state.IsBaselined(ctx, streamID)
	if err != nil {
		return fmt.Errorf("check baseline: %w", err)
	}
	baselineMode := !baselined

	var failed atomic.Int64
	var checked atomic.Int64
	var newDocuments atomic.Int64
	var sent atomic.Int64

	fmt.Printf("Lessons: %d\n", len(lessonIDs))

	var tgMu sync.Mutex

	logs := make([]string, len(lessonIDs))

	g, gctx := errgroup.WithContext(ctx)
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

			documents, err := client.GetLessonDocuments(gctx, id)
			if err != nil {
				failed.Add(1)
				printf("  Lesson ID: %s\n", id)
				printf("  ERROR: %v\n\n", err)
				logs[i] = logBuf.String()
				return nil
			}

			for _, document := range documents {
				checked.Add(1)

				name := strings.TrimPrefix(document.Name, "Занятия на ")
				name = strings.TrimSuffix(name, ".doc")

				data, err := client.DownloadDocument(gctx, document)
				if err != nil {
					failed.Add(1)
					printf("  Lesson ID: %s\n", document.ID)
					printf("  Document: %s\n", name)
					printf("  URL: %s\n", document.URL)
					printf("  ERROR Download: %v\n\n", err)
					continue
				}

				fileSize := float64(len(data)) / 1024

				if baselineMode {
					if err := state.MarkProcessed(gctx, document.URL, data); err != nil {
						failed.Add(1)
						printf("  ERROR Redis MarkProcessed (baseline): %v\n\n", err)
						continue
					}
					printf("  BASELINE: %s\n", name)
					continue
				}

				changed, err := state.HasChanged(gctx, document.URL, data)
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

				jpgPath, err := converter.ConvertToJPG(gctx, data, document.Name)
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
				err = telegramClient.SendPhoto(targetChatID, jpgPath, caption)
				tgMu.Unlock()

				if err != nil {
					printf("  ERROR Telegram: %v\n\n", err)
					continue
				}

				sent.Add(1)

				if err := notifyMaterial(document.Name, document.URL, document.ID, lessonURL, fileSize, checked.Load(), newDocuments.Load(), chatID); err != nil {
					log.Printf("telegram log %s: %v", document.Name, err)
				}

				if err := state.MarkProcessed(gctx, document.URL, data); err != nil {
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

	if baselineMode {
		if failed.Load() == 0 {
			if err := state.MarkBaselined(ctx, streamID); err != nil {
				return fmt.Errorf("mark baselined: %w", err)
			}
			fmt.Println("Baseline completed, nothing sent")
		} else {
			fmt.Printf("Baseline incomplete (%d errors), will retry next tick\n", failed.Load())
		}
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
