// Домашнее задание №2 к уроку Go лучшие практики
// Логирование
package main

//Исходники задания для первого занятия у других групп https://github.com/t0pep0/GB_best_go1

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"lesson3/findfiles"
	"lesson3/listdirectory"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//type TargetFile struct {
//	Path string
//	Name string
//}

//type FileList map[string]TargetFile

//type FileInfo interface {
//	os.FileInfo
//	Path() string
//}

type fileInfo struct {
	os.FileInfo
	path string
}

func (fi fileInfo) Path() string {
	return fi.path
}

//type SearchData struct {
//	sync.Mutex
//	Depth          int
//	Current        int
//	LastSignalType os.Signal
//	WaitCh         *chan struct{}
//}

type FileSearcher struct {
	logger       *zap.Logger
	fileSearcher *findfiles.FileSearcher
}

func NewFileSearcher(logger *zap.Logger) *FileSearcher {
	return &FileSearcher{
		logger: logger,
	}
}

/*
func (f *FileSearcher) ListDirectory(ctx context.Context, dir string, data *SearchData) ([]FileInfo, error) {
	// Ограничить глубину поиска заданным числом, по SIGUSR2 увеличить глубину поиска на +2
	*data.waitCh <- struct{}{}
	select {
	case <-ctx.Done():
		f.logger.Info("context is done, skipping dir", zap.String("dir", dir))
		return nil, nil
	default:
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		//По SIGUSR2 увеличить глубину поиска на +2
		time.Sleep(time.Second * 30) // Добавить общий таймаут на работу парсера
		switch data.lastSignalType {
		case syscall.SIGUSR1:
			fmt.Printf("\nDirectory: %s, Depth: %d", dir, data.depth)
			f.logger.Info("input syscall.SIGUSR1")
		case syscall.SIGUSR2:
			data.Lock()
			data.depth += 2
			data.Unlock()
			f.logger.Info("input syscall.SIGUSR2")
		}
		var result []FileInfo
		res, err := os.ReadDir(dir)
		if err != nil {
			f.logger.Error("error reading dir", zap.Error(err),
				zap.String("dir", dir))
			return nil, err
		}
		for _, entry := range res {
			data.current = 0
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				fmt.Println(data.current, data.depth, path)
				if data.current < data.depth {
					data.current++
					child, err := f.ListDirectory(ctx, path, data) //Дополнительно: вынести в горутину
					if err != nil {
						f.logger.Error("error reading subdirectory", zap.Error(err),
							zap.String("path", path))
						return nil, err
					}
					result = append(result, child...)
				}
			} else {
				info, err := entry.Info()
				if err != nil {
					f.logger.Error("error reading file.Info", zap.Error(err),
						zap.String("path", path))
					return nil, err
				}
				result = append(result, fileInfo{info, path})
			}
		}
		return result, nil
	}
}

func (f *FileSearcher) FindFiles(ctx context.Context, ext string, data *SearchData) (FileList, error) {
	wd, err := os.Getwd()
	if err != nil {
		f.logger.Error("Could not get work directory", zap.Error(err))
		return nil, err
	}
	files, err := f.ListDirectory(ctx, wd, data)
	if err != nil {
		f.logger.Error("Error not on get file list", zap.Error(err))
		if len(files) == 0 {
			return nil, err
		}
		f.logger.Warn("Error not on get path of file list", zap.Error(err))
	}
	fl := make(FileList, len(files))
	for _, file := range files {
		fileExt := filepath.Ext(file.Name())
		f.logger.Debug("Compare extenstions", zap.String("target_ext", ext),
			zap.String("current", fileExt))
		if fileExt == ext {
			fl[file.Name()] = TargetFile{
				Name: file.Name(),
				Path: file.Path(),
			}
		}
	}
	return fl, err
}
*/
var (
	GitHash   = ""
	BuildTime = ""
	Version   = ""
)

func main() {
	const (
		wantExt = ".go"
		//		development = "DEVELOPMENT"
		production = "PRODUCTION"
		env        = "ENV"
	)
	var logger *zap.Logger
	curEnv := os.Getenv(env) // Получаем текущее окружение
	var err error
	if curEnv == production {
		//logger, err = zap.NewProduction()
		logCfg := zap.NewProductionConfig()
		logCfg.OutputPaths = []string{"stderr"}
		logger, err = logCfg.Build()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal("Failed to initialize logger", err)
	}
	logger.Info("starting", zap.Int("pid", os.Getgid()),
		zap.String("commit_hash", GitHash), zap.String("BuildTime", BuildTime),
		zap.String("version", Version))
	//defer logger.Sync() // Сбрасываем все содержание текущего буфера в место, куда это все нужно запистать

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	waitCh := make(chan struct{})
	Data := listdirectory.SearchData{ // Вопрос: можно ли этот ввод данных перенести в ListDirectory.go?
		// Думаю, что ошибка возникает из-за передачи этой структуры.
		//Или ввод данных должен быть всегда в main.go? И кк это отразится на второй горутине в main?
		Depth:  2,
		WaitCh: &waitCh,
	}
	l := NewFileSearcher(logger)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	// TODO: Как проверить прохождение сигнала syscall.SIGUSR1 или syscall.SIGUSR2
	//syscall.Kill(syscall.Getpid(), syscall.SIGUSR1) // Пробую передать системный сигнал
	go func() {
		defer close(waitCh)
		res, err := l.fileSearcher.FindFiles(ctx, wantExt, &Data)
		if err != nil {
			logger.Error("Error on search", zap.Error(err))
			//log.Printf("Error on search: %v\n", err)
			os.Exit(1)
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
		}
	}()

	go func() {
		<-sigCh
		logger.Info("Signal received, terminate...")
		// log.Println("Signal received, terminate...")
		cancel()
		signalType := <-sigCh // При создании канал находится в постоянном ожидании приема системных сигналов?
		Data.Lock()
		Data.LastSignalType = signalType // безопасная запись в структуру, потому что другая горутина тоже пишет в структуру data, но в другие поля
		Data.Unlock()

		switch signalType { // Обработка принятых системных сигналов. Как сгенерировать пользовательский сигнал SIGUSR1 и SIGUSR2 ? Я не нашел информацию
		case syscall.SIGUSR1:
			log.Println("INPUT SIGUSR1: display current directory and current search depth") // Обработать сигнал SIGUSR1
		case syscall.SIGUSR2:
			log.Println("INPUT SIGUSR2: search depth will be increased (+2)") // Обработать сигнал SIGUSR2
		default:
			log.Println("Signal received, terminate...") // Текстовая информация та, которая соответствует всем каналам кроме SIGUSR1 и SIGUSR2 ? Или нужно изменить?
			cancel()
		}
	}()
	//Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	logger.Info("Done")
}
