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
	o            Options
	PrintMsgFunc func(msg string)
}

func NewWorker(opts ...Option) *Worker {
	worker := new(Worker)
	options := newOptions(opts...)

	worker.o = options
	worker.PrintMsgFunc = func(msg string) { log.Println(msg) }

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

	err := filepath.Walk(w.o.SourcePath, func(oldPath string, info os.FileInfo, err error) error {
		if err != nil {
			w.PrintMsgFunc(fmt.Sprintf("failed to access file. path:%s err:%v\n", oldPath, err))
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(path.Ext(oldPath))
		if ext != ".mp3" && ext != ".flac" {
			w.PrintMsgFunc(fmt.Sprintf("Not Supported ext: %s path: %s", ext, oldPath))
			return nil
		}

		f, err := os.Open(oldPath)
		if err != nil {
			w.PrintMsgFunc(fmt.Sprintf("failed to open file. path:%s err:%v\n", oldPath, err))
			return err
		}
		defer f.Close()

		meta, err := tag.ReadFrom(f)
		if err != nil {
			w.PrintMsgFunc(fmt.Sprintf("failed to read tag from file. path:%s err:%v\n", oldPath, err))
			return nil
		}

		dstDir := fmt.Sprintf("%s/%s/%s", w.o.TargetPath, Escape(meta.Artist()), Escape(meta.Album()))
		err = os.MkdirAll(dstDir, 0777)
		if err != nil {
			panic(err)
		}
		dstFilePath := fmt.Sprintf("%s/%s%s", dstDir, Escape(meta.Title()), ext)
		if w.o.DryRun {
			w.PrintMsgFunc(fmt.Sprintf("DRYRUN: %s => %s\n", oldPath, dstFilePath))
			return nil
		}
		f.Seek(0, 0)
		if _, err := os.Stat(dstFilePath); !os.IsNotExist(err) {
			// old same name file exist!
			of, err := os.Open(dstFilePath)
			if err != nil {
				w.PrintMsgFunc(fmt.Sprintf("failed to open file. path:%s meta.Artist:%s meta.Album:%s err:%v\n", dstFilePath, meta.Artist(), meta.Album(), err))
				return err
			}
			defer of.Close()

			if w.o.DuplicationAutoReplace {
				w.PrintMsgFunc(fmt.Sprintf("auto replace old file %s", dstFilePath))
			} else {
				oldMeta, err := tag.ReadFrom(of)
				if err != nil {
					w.PrintMsgFunc(fmt.Sprintf("failed to read tag from file. path:%s err:%v\n", dstFilePath, err))
					return nil
				}
				diff, equal := messagediff.PrettyDiff(oldMeta, meta)
				if equal {
					w.PrintMsgFunc(fmt.Sprintf("same file and same meta info exist! ignore: %s => %s", oldPath, dstFilePath))
					return nil
				}
				w.PrintMsgFunc(fmt.Sprintf("[ATTENTION]same path\n src: %s\ndest: %s\nold info:%s\ndiff:\n%s\n\nPlease check replace[y/n]:", oldPath, dstFilePath, spew.Sdump(oldMeta), diff))
				var replace string
				n, _ := fmt.Scanln(&replace)
				replace = strings.ToLower(replace)
				if n != 1 || replace == "n" {
					w.PrintMsgFunc(fmt.Sprintf("Not replace"))
					return nil
				}
			}
		}

		dstFile, err := os.OpenFile(dstFilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			w.PrintMsgFunc(fmt.Sprintf("failed to open and create file. path:%s meta.Artist:%s meta.Album:%s err:%v\n", dstFilePath, meta.Artist(), meta.Album(), err))
			return nil
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, f); err != nil {
			w.PrintMsgFunc(fmt.Sprintf("failed to copy. err:%v\n", err))
			return err
		}

		if w.o.DeleteSrc {
			os.Remove(oldPath)
		}

		return nil
	})

	if err != nil {
		w.PrintMsgFunc(fmt.Sprintf("walk failed! err:%v\n", err))
	}
}
