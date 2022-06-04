package listdirectory

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

/*
type TargetFile struct {
	Path string
	Name string
}
*/

/*
type FileList map[string]TargetFile
*/

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

type SearchData struct {
	sync.Mutex
	Depth          int
	Current        int
	LastSignalType os.Signal
	WaitCh         *chan struct{}
}

type FileSearcher struct {
	logger *zap.Logger
}

/*
func NewFileSearcher(logger *zap.Logger) *FileSearcher {
	return &FileSearcher{
		logger: logger,
	}
}
*/

func (f *FileSearcher) ListDirectory(ctx context.Context, dir string, data *SearchData) ([]FileInfo, error) {
	// Ограничить глубину поиска заданным числом, по SIGUSR2 увеличить глубину поиска на +2
	*data.WaitCh <- struct{}{}
	select {
	case <-ctx.Done():
		f.logger.Info("context is done, skipping dir", zap.String("dir", dir))
		return nil, nil
	default:
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		//По SIGUSR2 увеличить глубину поиска на +2
		time.Sleep(time.Second * 30) // Добавить общий таймаут на работу парсера
		switch data.LastSignalType {
		case syscall.SIGUSR1:
			fmt.Printf("\nDirectory: %s, Depth: %d", dir, data.Depth)
			f.logger.Info("input syscall.SIGUSR1")
		case syscall.SIGUSR2:
			data.Lock()
			data.Depth += 2
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
			data.Current = 0
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				fmt.Println(data.Current, data.Depth, path)
				if data.Current < data.Depth {
					data.Current++
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
		fmt.Println("I am end ListDirectory") // Убрать после отладки !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		return result, nil
	}
}
