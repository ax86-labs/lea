package watcher

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/PizenLabs/lea/internal/parser/contracts"
	"github.com/PizenLabs/lea/internal/parser/golang"
	storage "github.com/PizenLabs/lea/internal/storage/contracts"
	"github.com/PizenLabs/lea/internal/workspace/ignore"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	store  storage.Store
	parser contracts.Parser
	root   string
}

func NewWatcher(store storage.Store, root string) *Watcher {
	return &Watcher{
		store:  store,
		parser: golang.NewParser(),
		root:   root,
	}
}

func (w *Watcher) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Recursively add directories to watch
	matcher := ignore.NewMatcher(w.root)
	err = filepath.WalkDir(w.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if matcher.ShouldSkipDir(path, entry) {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Watching for changes in %s...\n", w.root)

	// Debounce timer map to avoid multiple rapid updates for the same file
	timers := make(map[string]*time.Timer)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if t, ok := timers[event.Name]; ok {
					t.Stop()
				}
				timers[event.Name] = time.AfterFunc(500*time.Millisecond, func() {
					w.handleUpdate(ctx, event.Name)
				})
			} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				w.handleDelete(ctx, event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("error: %v", err)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *Watcher) handleUpdate(ctx context.Context, path string) {
	fmt.Printf("File changed: %s, updating index...\n", path)

	// Surgical update: delete old nodes for this file and re-parse
	if err := w.store.DeleteByFile(ctx, path); err != nil {
		log.Printf("Error deleting old nodes for %s: %v", path, err)
		return
	}

	nodes, edges, err := w.parser.ParseFile(ctx, path)
	if err != nil {
		log.Printf("Error parsing %s: %v", path, err)
		return
	}

	callEdges, callErr := w.parser.ExtractCalls(ctx, path)
	if callErr != nil {
		log.Printf("Error extracting calls for %s: %v", path, callErr)
		return
	}

	flowEdges, flowErr := w.parser.ExtractControlFlow(ctx, path)
	if flowErr != nil {
		log.Printf("Error extracting flow for %s: %v", path, flowErr)
		return
	}

	edges = append(edges, callEdges...)
	edges = append(edges, flowEdges...)

	if err := w.store.SaveGraph(ctx, nodes, edges); err != nil {
		log.Printf("Error saving graph for %s: %v", path, err)
		return
	}

	fmt.Printf("Updated %s\n", path)
}

func (w *Watcher) handleDelete(ctx context.Context, path string) {
	fmt.Printf("File deleted: %s, removing from index...\n", path)
	if err := w.store.DeleteByFile(ctx, path); err != nil {
		log.Printf("Error deleting nodes for %s: %v", path, err)
	}
}
