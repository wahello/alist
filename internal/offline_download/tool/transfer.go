package tool

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/stream"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xhofe/tache"
	"path/filepath"
	"sync"
)

type TransferTask struct {
	tache.Base
	file       File
	dstDirPath string
	wg         *sync.WaitGroup
	tempDir    string
}

func (t *TransferTask) Run() error {
	defer t.wg.Done()
	// check dstDir again
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(t.dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	mimetype := utils.GetMimeType(t.file.Path)
	rc, err := t.file.GetReadCloser()
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", t.file.Path)
	}
	s := &stream.FileStream{
		Ctx: nil,
		Obj: &model.Object{
			Name:     filepath.Base(t.file.Path),
			Size:     t.file.Size,
			Modified: t.file.Modified,
			IsFolder: false,
		},
		Reader:   rc,
		Mimetype: mimetype,
		Closers:  utils.NewClosers(rc),
	}
	relDir, err := filepath.Rel(t.tempDir, filepath.Dir(t.file.Path))
	if err != nil {
		log.Errorf("find relation directory error: %v", err)
	}
	newDistDir := filepath.Join(dstDirActualPath, relDir)
	return op.Put(t.Ctx(), storage, newDistDir, s, t.SetProgress)
}

func (t *TransferTask) GetName() string {
	return fmt.Sprintf("transfer %s to [%s]", t.file.Path, t.dstDirPath)
}

func (t *TransferTask) GetStatus() string {
	return "transferring"
}

var (
	TransferTaskManager *tache.Manager[*TransferTask] = tache.NewManager[*TransferTask]()
)
