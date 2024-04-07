package server

import (
	"context"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	ootPurgeInterval = 5 * time.Minute
	ootMaxLifetime   = 10 * time.Minute
)

// oot represents an Other Object Type
type oot struct {
	evt        s3Event
	appendTime time.Time
}

type objectTemporizer struct {
	baseDirChan       chan string
	ootChan           chan s3Event
	cache             *cache
	unassignedObjects map[string][]oot
	objectsLock       sync.Mutex
}

func newObjectTemporizer(baseDirChan chan string, ootChan chan s3Event, cache *cache) *objectTemporizer {
	return &objectTemporizer{
		baseDirChan:       baseDirChan,
		ootChan:           ootChan,
		cache:             cache,
		unassignedObjects: make(map[string][]oot),
	}
}

func (op *objectTemporizer) goTemporize(ctx context.Context) {
	go func() {
		purgeTicker := time.NewTicker(ootPurgeInterval)
		defer purgeTicker.Stop()

		for ctx.Err() == nil {
			select {
			case event, ok := <-op.ootChan:
				if !ok {
					return
				}

				op.objectsLock.Lock()
				op.handleEvent(ctx, event)
				op.objectsLock.Unlock()
			case baseDir, ok := <-op.baseDirChan:
				if !ok {
					return
				}

				op.objectsLock.Lock()
				op.handleBaseDir(ctx, baseDir)
				op.objectsLock.Unlock()
			case <-purgeTicker.C:
				op.objectsLock.Lock()
				op.purge(time.Now())
				op.objectsLock.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (op *objectTemporizer) handleEvent(ctx context.Context, event s3Event) {
	objDir := path.Dir(event.ObjectKey)

	if match, baseDir := op.cache.matchesEntry(event.Bucket, objDir); match {
		event.baseDir = baseDir
		go op.cache.handleEvent(ctx, event)
	} else {
		op.unassignedObjects[objDir] = append(op.unassignedObjects[objDir], oot{event, time.Now()})
	}
}

func (op *objectTemporizer) handleBaseDir(ctx context.Context, baseDir string) {
	for dir, oots := range op.unassignedObjects {
		if strings.HasPrefix(dir, baseDir) {
			go op.signalObjects(ctx, baseDir, oots)
			delete(op.unassignedObjects, dir)
		}
	}
}

func (op *objectTemporizer) signalObjects(ctx context.Context, baseDir string, oots []oot) {
	for _, oot := range oots {
		evt := oot.evt
		evt.baseDir = baseDir

		op.cache.handleEvent(ctx, evt)
	}
}

func (op *objectTemporizer) purge(now time.Time) {
	for dir, oots := range op.unassignedObjects {
		i := 0

		for _, oot := range oots {
			if now.Sub(oot.appendTime) > ootMaxLifetime {
				continue
			}

			oots[i] = oot
			i++
		}

		if i == 0 {
			delete(op.unassignedObjects, dir)
		} else {
			op.unassignedObjects[dir] = oots[:i]
		}
	}
}
