package findfiles

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
	"log"
	"os"
	"testing"
)

func NewFileSearcher(logger *zap.Logger) *FileSearcher {
	return &FileSearcher{
		logger: logger,
	}
}

// TestFindFiles function findfiles.FindFiles
func TestFindFiles(t *testing.T) {
	var logger *zap.Logger
	var (
		GitHash   = ""
		BuildTime = ""
		Version   = ""
	)
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
	loggerCurrent := NewFileSearcher(logger) // не находит, потому что это метод
	ctx := context.Background()
	const wantExt = ".go"
	var nestingDepth = 2
	res, _ := loggerCurrent.FindFiles(ctx, wantExt, nestingDepth, nil)
	resTest := map[string]TargetFile{
		"main.go": {
			"/home/user/go_best_practics/lesson4/cmd/main.go",
			"main.go",
		},
		"file1.go": {
			"/home/user/go_best_practics/lesson4/subdir1/file1.go",
			"file1.go",
		},
	}
	Convey("Test map key", func() {
		So(res, ShouldContainKey, "main.go") // для мап'ов - должен содержать ключ
	})
	Convey("Test map key", func() {
		So(res, ShouldContainKey, "file1.go") // для мап'ов - должен содержать ключ
	})
	Convey("Test map", func() {
		So(res, ShouldResemble, resTest) // "глубокое" сравнение массивов, слайсов, мап'ов и структур
	})
}
