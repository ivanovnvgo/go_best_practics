package findfiles

import (
	"context"
	"go.uber.org/zap"
	"lesson3/listdirectory"
	"os"
	"path/filepath"
)

type TargetFile struct {
	Path string
	Name string
}

type FileList map[string]TargetFile

/*
type FileInfo interface {
	os.FileInfo
	Path() string
}
*/
/*
type fileInfo struct {
	os.FileInfo
	path string
}
*/

/*
func (fi fileInfo) Path() string {
	return fi.path
}
*/
/*
type SearchData struct {
	sync.Mutex
	depth          int
	current        int
	lastSignalType os.Signal
	waitCh         *chan struct{}
}
*/

type FileSearcher struct {
	Logger       *zap.Logger
	FileSearcher *listdirectory.FileSearcher
}

/*
func NewFileSearcher(logger *zap.Logger) *FileSearcher {
	return &FileSearcher{
		logger: logger,
	}
}
*/

func (f *FileSearcher) FindFiles(ctx context.Context, ext string, Data *listdirectory.SearchData) (FileList, error) {
	wd, err := os.Getwd()
	if err != nil {
		f.Logger.Error("Could not get work directory", zap.Error(err))
		return nil, err
	}
	files, err := f.FileSearcher.ListDirectory(ctx, wd, Data)
	if err != nil {
		f.Logger.Error("Error not on get file list", zap.Error(err))
		if len(files) == 0 {
			return nil, err
		}
		f.Logger.Warn("Error not on get path of file list", zap.Error(err))
	}
	fl := make(FileList, len(files))
	for _, file := range files {
		fileExt := filepath.Ext(file.Name())
		f.Logger.Debug("Compare extenstions", zap.String("target_ext", ext),
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
