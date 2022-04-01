// Домашнее задание №2 к уроку Go лучшие практики
// Логирование
// Здесь недоделанное домашнее задание к уроку №1, напишу к нему логи, потом снова вернусь к обработке сигналов в hw1.go
package main

//Исходники задания для первого занятия у других групп https://github.com/t0pep0/GB_best_go1

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type TargetFile struct {
	Path string
	Name string
}

type FileList map[string]TargetFile

type FileInfo interface {
	os.FileInfo
	Path() string
}

type fileInfo struct {
	os.FileInfo
	path string
}

func (fi fileInfo) Path() string {
	return fi.path
}

type FileSearcher struct {
	logger *zap.Logger
}

func NewFileSearcher(logger *zap.Logger) *FileSearcher {
	return &FileSearcher{
		logger: logger,
	}
}

func (f *FileSearcher) listDirectory(ctx context.Context, dir string) ([]FileInfo, error) {
	// Ограничить глубину поиска заданым числом, по SIGUSR2 увеличить глубину поиска на +2
	select {
	case <-ctx.Done():
		f.logger.Info("context is done, skipping dir", zap.String("dir", dir))
		return nil, nil
	default:
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		time.Sleep(time.Second * 10) // Добавить общий таймаут на работу парсера
		var result []FileInfo
		res, err := os.ReadDir(dir)
		if err != nil {
			f.logger.Error("error reading dir", zap.Error(err),
				zap.String("dir", dir))
			return nil, err
		}
		for _, entry := range res {
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				child, err := f.listDirectory(ctx, path) //Дополнительно: вынести в горутину
				if err != nil {
					return result, err
				}
				result = append(result, child...)
			} else {
				info, err := entry.Info()
				if err != nil {
					return result, err
				}
				result = append(result, fileInfo{info, path})
			}
		}
		return result, nil
	}
}

func (f *FileSearcher) FindFiles(ctx context.Context, ext string) (FileList, error) {
	wd, err := os.Getwd()
	if err != nil {
		f.logger.Error("Could not get work directory", zap.Error(err))
		return nil, err
	}
	files, err := f.listDirectory(ctx, wd)
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

var (
	GitHash   = ""
	BuildTime = ""
	Version   = ""
)

func main() {
	const searchDepth = 2 //Ограничить глубину поиска заданным числом

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
		logger, err = zap.NewProduction()
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
	defer logger.Sync() // Сбрасываем все содержание текущего буфера в место, куда это все нужно запистать

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	select {
	case sig := <-sigCh: // TODO: проверить прохождение сигнала syscall.SIGUSR1 или syscall.SIGUSR2
		depth := searchDepth + 2                                             // TODO: новая глубина поиска, передать в ListDirectory
		fmt.Printf("\ninput os signal: %s\ndepth searches %d\n", sig, depth) // Удалить после отладки !!!
	}
	//Обработать сигнал SIGUSR1
	waitCh := make(chan struct{})
	fileSearcher := NewFileSearcher(logger)
	//osSignalChan := make(chan os.Signal) // Обработать сигнал SIGUSR1
	//signal.Notify(osSignalChan,
	//	syscall.SIGINT,
	//	syscall.SIGTERM)
	//sig := <-osSignalChan
	//log.Printf("got signal %q", sig.String())

	go func() {
		res, err := fileSearcher.FindFiles(ctx, wantExt)
		if err != nil {
			logger.Error("Error on search", zap.Error(err))
			//log.Printf("Error on search: %v\n", err)
			os.Exit(1)
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
		}
		waitCh <- struct{}{}
	}()
	go func() {
		<-sigCh
		logger.Info("Signal received, terminate...")
		// log.Println("Signal received, terminate...")
		cancel()
	}()
	//Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	logger.Info("Done")
}
