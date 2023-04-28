package storage

import (
	"errors"
	"fmt"
	arangodriver "github.com/arangodb/go-driver"
	"github.com/clusterpedia-io/clusterpedia/pkg/storage"
	"gorm.io/gorm"
	"io"
	genericstorage "k8s.io/apiserver/pkg/storage"
	"net"
	"os"
	"syscall"
)

func InterpretResourceDBError(cluster, name string, err error) error {
	if err == nil {
		return nil
	}

	return InterpretDBError(fmt.Sprintf("%s/%s", cluster, name), err)
}

func InterpretDBError(key string, err error) error {
	if err == nil {
		return nil
	}

	if arangoErr, ok := err.(arangodriver.ArangoError); ok {
		if arangoErr.ErrorNum == arangodriver.ErrArangoUniqueConstraintViolated {
			return genericstorage.NewKeyExistsError(key, 0)
		}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return genericstorage.NewKeyNotFoundError(key, 0)
	}

	if _, isNetError := err.(net.Error); isNetError ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		os.IsTimeout(err) ||
		errors.Is(err, os.ErrDeadlineExceeded) ||
		errors.Is(err, syscall.ECONNREFUSED) {
		return storage.NewRecoverableException(err)
	}

	return err
}
