// Лучшие практики разработки Go-приложений
// Урок 2. Обработка ошибок сторонних сервисов и сигналов операционной системы
// Урок 3. Логирование
// Урок 4. Продвинутые практики тестирования
// Урок 5. Принципы структурирования Go-приложений
// Урок 6. Линтеры: продвинутый уровень
// Урок 7. Сборка приложений и автоматизация повторяющихся действий
package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	findfiles "lesson4/internal/findfiles"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// FileSearcher Начало блока логирования
type FileSearcher struct {
	logger    *zap.Logger
	findfiles findfiles.FileSearcher // Это правильное решение?????
}

func NewFileSearcher(logger *zap.Logger) *FileSearcher {
	return &FileSearcher{
		logger: logger,
	}
}

var (
	GitHash   = ""
	BuildTime = ""
	Version   = ""
)

// Конец блока логирования

func main() {
	const wantExt = ".go"

	var nestingDepth = 2 // задаем глубину вложенности при поиске нужных файлов, максимальное значение 4

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)

	var wg sync.WaitGroup
	wg.Add(1)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	// Обработать сигнал SIGUSR1
	waitCh := make(chan struct{})

	// Создаем структуру dataInput тип SearchData для передачи данных в различные функции
	dataInput := findfiles.DataInput{Ch: &waitCh}
	// Блок объявления логирования
	var logger *zap.Logger
	var err error
	logger, err = zap.NewProduction()
	logCfg := zap.NewProductionConfig()
	logCfg.OutputPaths = []string{"stderr"}
	logger, err = logCfg.Build()
	if err != nil {
		log.Fatal("Failed to initialize logger", err)
	}
	logger.Info("starting", zap.Int("pid", os.Getpid()),
		zap.String("commit_hash", GitHash), zap.String("BuildTime", BuildTime),
		zap.String("version", Version))
	loggerCurrent := NewFileSearcher(logger)
	// Сброс содержимого текущего буфера в место, куда все это содержимое нужно было записать
	defer logger.Sync()
	// Конец блока логирования

	go func() {
		defer wg.Done()
		res, err := loggerCurrent.findfiles.FindFiles(ctx, wantExt, nestingDepth, &dataInput)
		if err != nil {
			logger.Error("Error of search", zap.Error(err))
			os.Exit(1)
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
			fmt.Println(f)
		}
		waitCh <- struct{}{}
	}()

	go func() {
		defer wg.Done()
		signalType, ok := <-sigCh // При создании канал находится в постоянном ожидании приема системных сигналов?
		if ok != true {
			logger.Error("Channel for read os signals is close")
		}
		switch signalType { // Обработка принятых системных сигналов.
		// Как сгенерировать пользовательский сигнал SIGUSR1 и SIGUSR2 ? Я не нашел информацию
		case syscall.SIGUSR1:
			log.Println("INPUT SIGUSR1: display current directory and current search depth") // Обработать сигнал SIGUSR1
		case syscall.SIGUSR2:
			log.Println("INPUT SIGUSR2: search depth will be increased (+2)") // Обработать сигнал SIGUSR2
		default:
			log.Println("Signal received, terminate...") // Текстовая информация та, которая соответствует всем каналам кроме SIGUSR1 и SIGUSR2
		}
		cancel()
	}()

	// Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	wg.Wait()
	logger.Info("Done")
}
