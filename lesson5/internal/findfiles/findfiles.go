package findfiles

import (
	"context"
	"fmt"
	"os"
	"path"
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

type DataInput struct { // Добавлен тип DataInput структура
	signalOs os.Signal
	Ch       *chan struct{}
}

// ListDirectory - рекурсивная функция, принимает контекст и текущую директорию,
//возвращает список вложенных файлов и ошибку

func ListDirectory(ctx context.Context, dir string, depthChildDir int, dataInput *DataInput) ([]FileInfo, error) {
	// Ограничить глубину поиска заданным числом, по SIGUSR2 увеличить глубину поиска на +2
	// chanChildDir принимает значение вложенной директории child (при рекурсивном вызове ListDirectory)
	chanChildDir := make(chan []FileInfo, 1)
	// chanChildDirErr принимает значение ошибки err (при рекурсивном вызове ListDirectory)
	chanChildDirErr := make(chan error, 1)
	defer close(chanChildDir)
	defer close(chanChildDirErr)
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		time.Sleep(time.Second * 5)
		switch dataInput.signalOs { // Проверка принятого системного сигнала
		//По SIGUSR1 вывести текущую директорию и текущую глубину поиска
		case syscall.SIGUSR1:
			fmt.Printf("\nDirectory: %s, Depth: %d\n", dir, depthChildDir)
		case syscall.SIGUSR2:
			// По SIGUSR2 увеличить глубину поиска на +2
			depthChildDir += 2
		}

		var result []FileInfo
		res, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range res {
			if depthChildDir == 0 { // Вниз глубже заданной глубины поиска не спускаемся
				break
			}
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				go func() {
					//Дополнительно: вынести в горутину
					child, err := ListDirectory(ctx, path, depthChildDir-1, dataInput)
					chanChildDir <- child
					chanChildDirErr <- err
					//fmt.Println(depthChildDir)
				}()
				//fmt.Printf("Type of chanChildDir: %T, Type of chanChildDirErr: %T\n", chanChildDir, chanChildDirErr)
				child := <-chanChildDir
				err = <-chanChildDirErr
				//fmt.Printf("Type of child: %T, Type of err: %T\n", child, err)
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

// FindFiles - - рекурсивная функция, принимает контекст, расширение файла, глубину поиска и структуру,
//возвращает список файлов (map), соответствующих принятому расширению, и ошибку

func FindFiles(ctx context.Context, ext string, depth int, dataInput *DataInput) (FileList, error) {
	wd, err := os.Getwd() // получили текущую директорию wd
	if err != nil {
		return nil, err
	}
	newWd := path.Dir(wd) // получили директорию на 1 уровень ниже текущей wd
	files, err := ListDirectory(ctx, newWd, depth, dataInput)
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
