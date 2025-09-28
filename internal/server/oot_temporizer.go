package server

import (
	"context"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

const (
	ootPurgeInterval = 5 * time.Minute
	ootMaxLifetime   = 10 * time.Minute
)

// oot represents an Other Object Type.
type oot struct {
	evt        s3Event
	appendTime time.Time
}

type objectTemporizer struct {
	baseDirChan       chan string
	temporizationChan chan s3Event
	cache             *cache
	productsCfg       config.Products
	unassignedObjects map[string][]oot
	objectsLock       sync.Mutex
}

func newObjectTemporizer(baseDirChan chan string, ootChan chan s3Event, cache *cache, productsCfg config.Products) *objectTemporizer {
	return &objectTemporizer{
		baseDirChan:       baseDirChan,
		temporizationChan: ootChan,
		cache:             cache,
		productsCfg:       productsCfg,
		unassignedObjects: make(map[string][]oot),
	}
}

func (op *objectTemporizer) goTemporize(ctx context.Context) {
	go func() {
		purgeTicker := time.NewTicker(ootPurgeInterval)
		defer purgeTicker.Stop()

		for ctx.Err() == nil {
			select {
			case event, ok := <-op.temporizationChan:
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
		evt, ok := op.computeEvent(event, baseDir)
		if ok {
			go op.cache.handleEvent(ctx, evt)
		}
	} else {
		op.unassignedObjects[objDir] = append(op.unassignedObjects[objDir], oot{event, time.Now()})
	}
}

func (op *objectTemporizer) handleBaseDir(ctx context.Context, baseDir string) {
	for dir, oots := range op.unassignedObjects {
		if dir == baseDir || strings.HasPrefix(dir, baseDir+"/") {
			go op.signalObjects(ctx, baseDir, oots)

			delete(op.unassignedObjects, dir)
		}
	}
}

func (op *objectTemporizer) signalObjects(ctx context.Context, baseDir string, oots []oot) {
	for _, oot := range oots {
		evt, ok := op.computeEvent(oot.evt, baseDir)
		if !ok {
			continue
		}

		op.cache.handleEvent(ctx, evt)
	}
}

func (op *objectTemporizer) computeEvent(ootEvt s3Event, baseDir string) (s3Event, bool) {
	if ootEvt.ObjectType == types.ObjectNotYetAssigned {
		if !strings.HasPrefix(ootEvt.ObjectKey, baseDir) {
			return s3Event{}, false
		}

		objKeyWithoutBaseDir := strings.TrimPrefix(ootEvt.ObjectKey, baseDir)
		if op.productsCfg.TargetRelativeRgx.MatchString(objKeyWithoutBaseDir) {
			ootEvt.ObjectType = types.ObjectTarget
		} else {
			return s3Event{}, false
		}
	}

	ootEvt.baseDir = baseDir

	return ootEvt, true
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
