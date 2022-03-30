// Домашнее задание №1 к уроку Go лучшие практики
// При отправке парсеру сигнала SIGUSR1 он должен увеличить глубину поиска на 2
// Добавить общий таймаут на работу парсера +
// По SIGUSR1 вывести текущую директорию и текущую глубину поиска
// Ограничить глубину поиска заданым числом, по SIGUSR2 увеличить глубину поиска на +2
// Обработать сигнал SIGUSR1

package main

//Исходники задания для первого занятия у других групп https://github.com/t0pep0/GB_best_go - не работает((

import (
	"context"
	"fmt"
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

func ListDirectory(ctx context.Context, dir string) ([]FileInfo, error) {
	// Ограничить глубину поиска заданым числом, по SIGUSR2 увеличить глубину поиска на +2
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		time.Sleep(time.Second * 10) // Добавить общий таймаут на работу парсера
		var result []FileInfo
		res, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range res {
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				child, err := ListDirectory(ctx, path) //Дополнительно: вынести в горутину
				if err != nil {
					return nil, err
				}
				result = append(result, child...)
			} else {
				info, err := entry.Info()
				if err != nil {
					return nil, err
				}
				result = append(result, fileInfo{info, path})
			}
		}
		return result, nil
	}
}

func FindFiles(ctx context.Context, ext string) (FileList, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	files, err := ListDirectory(ctx, wd)
	if err != nil {
		return nil, err
	}
	fl := make(FileList, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			fl[file.Name()] = TargetFile{
				Name: file.Name(),
				Path: file.Path(),
			}
		}
	}
	return fl, nil
}

func main() {
	const searchDepth = 2 //Ограничить глубину поиска заданым числом
	const wantExt = ".go"
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

	//osSignalChan := make(chan os.Signal) // Обработать сигнал SIGUSR1
	//signal.Notify(osSignalChan,
	//	syscall.SIGINT,
	//	syscall.SIGTERM)
	//sig := <-osSignalChan
	//log.Printf("got signal %q", sig.String())

	go func() {
		res, err := FindFiles(ctx, wantExt)
		if err != nil {
			log.Printf("Error on search: %v\n", err)
			os.Exit(1)
		}
		for _, f := range res {
			fmt.Printf("\tName: %s\t\t Path: %s\n", f.Name, f.Path)
		}
		waitCh <- struct{}{}
	}()
	go func() {
		<-sigCh
		log.Println("Signal received, terminate...")
		cancel()
	}()
	//Дополнительно: Ожидание всех горутин перед завершением
	<-waitCh
	log.Println("Done")
}
