package tidy

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/dhowden/tag"
	"gopkg.in/d4l3k/messagediff.v1"
)

type Worker struct {
	o Options
}

func NewWorker(opts ...Option) *Worker {
	worker := new(Worker)
	options := newOptions(opts...)

	worker.o = options

	return worker
}

func (w *Worker) Tidy() {
	if w.o.Pattern == "" {
		w.defaultPattern()
	} else {
		panic("not support other pattern")
	}
}

func (w *Worker) defaultPattern() {
	if w.o.TargetPath == "" || w.o.TargetPath == w.o.SourcePath {
		log.Println("options fail")
		os.Exit(0)
	}

	fmt.Println("test")

	err := filepath.Walk(w.o.SourcePath, func(oldPath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("failed to access file. path:%s err:%v\n", oldPath, err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(path.Ext(oldPath))
		if ext != ".mp3" && ext != ".flac" {
			log.Printf("Not Supported ext: %s path: %s", ext, oldPath)
			return nil
		}

		f, err := os.Open(oldPath)
		if err != nil {
			log.Printf("failed to open file. path:%s err:%v\n", oldPath, err)
			return err
		}
		defer f.Close()

		meta, err := tag.ReadFrom(f)
		if err != nil {
			log.Printf("failed to read tag from file. path:%s err:%v\n", oldPath, err)
			return nil
		}

		dstDir := fmt.Sprintf("%s/%s/%s", w.o.TargetPath, meta.Artist(), meta.Album())
		f.Seek(0, 0)
		os.MkdirAll(dstDir, 0777)
		dstFilePath := fmt.Sprintf("%s/%s%s", dstDir, meta.Title(), ext)
		fmt.Println(dstFilePath)
		if _, err := os.Stat(dstFilePath); !os.IsNotExist(err) {
			// old same name file exist!
			of, err := os.Open(dstFilePath)
			if err != nil {
				log.Printf("failed to open file. path:%s err:%v\n", dstFilePath, err)
				return err
			}
			defer of.Close()

			oldMeta, err := tag.ReadFrom(of)
			if err != nil {
				log.Printf("failed to read tag from file. path:%s err:%v\n", dstFilePath, err)
				return nil
			}
			diff, equal := messagediff.PrettyDiff(oldMeta, meta)
			if equal {
				return nil
			}
			fmt.Printf("[ATTENTION]same path\n src: %s\ndest: %s\nold info:%s\ndiff:\n%s\n\nPlease check replace[y/n]:", oldPath, dstFilePath, spew.Sdump(oldMeta), diff)
			if w.o.DuplicationAutoReplace {
				log.Printf("auto replace old file")
			} else {
				var replace string
				n, _ := fmt.Scanln(&replace)
				replace = strings.ToLower(replace)
				if n != 1 || replace == "n" {
					log.Printf("Not replace")
					return nil
				}
			}
		}

		dstFile, err := os.OpenFile(dstFilePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			fmt.Printf("failed to open and create file. path:%s err:%v\n", dstFilePath, err)
			return nil
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, f); err != nil {
			log.Printf("failed to copy. err:%v\n", err)
			return err
		}

		if w.o.DeleteSrc {
			os.Remove(oldPath)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("walk failed! err:%v\n", err)
	}
}
