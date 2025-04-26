// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Contributors:
//	Fraunhofer AISEC

package testutil

import (
	"errors"
	"testing"

	"clouditor.io/clouditor/persistence"
	"clouditor.io/clouditor/persistence/gorm"
	"gorm.io/gorm/logger"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
)

// NewInMemoryStorage uses the Clouditor inmemory package to create a new in-memory storage that can be used for unit
// testing. The funcs varargs can be used to immediately execute storage operations on it.
func NewInMemoryStorage(t *testing.T, funcs ...func(s persistence.Storage)) (s persistence.Storage) {
	var err error

	s, err = gorm.NewStorage(
		gorm.WithInMemory(),
		gorm.WithMaxOpenConns(1),
		gorm.WithAdditionalAutoMigration(
			evaluation.EvaluationResult{},
			evaluation.Compliance{},
			collection.CollectionModule{},
			common.Evidence{},
			common.Error{},
			collection.ServiceConfiguration{}),
		// For storage debugging, set to `logger.Info`
		gorm.WithLogger(logger.Default.LogMode(logger.Silent)))
	if err != nil {
		t.Errorf("Could not initialize in-memory db: %v", err)
	}

	for _, f := range funcs {
		f(s)
	}

	return
}

// StorageWithError can be used to introduce various errors in a storage operation during unit testing.
type StorageWithError struct {
	CreateErr error
	SaveErr   error
	UpdateErr error
	GetErr    error
	ListErr   error
	CountErr  error
	DeleteErr error
}

func (s *StorageWithError) Create(_ interface{}) error                 { return s.CreateErr }
func (s *StorageWithError) Save(_ interface{}, _ ...interface{}) error { return s.SaveErr }
func (*StorageWithError) Update(_ interface{}, _ interface{}, _ ...interface{}) error {
	return nil
}
func (s *StorageWithError) Get(_ interface{}, _ ...interface{}) error { return s.GetErr }
func (s *StorageWithError) List(_ interface{}, _ string, _ bool, _ int, _ int, _ ...interface{}) error {
	return s.ListErr
}
func (s *StorageWithError) Count(_ interface{}, _ ...interface{}) (int64, error) {
	return 0, s.CountErr
}
func (s *StorageWithError) Delete(_ interface{}, _ ...interface{}) error { return s.DeleteErr }

// NewInMemoryStorageWithListError uses the Clouditor in-memory package to create a new in-memory storage that can be used for unit
// testing. The funcs varargs can be used to immediately execute storage operations on it.
func NewInMemoryStorageWithListError(t *testing.T, listErrorMsg string, funcs ...func(s persistence.Storage)) persistence.Storage {
	var err error
	s := &StorageWithListError{}
	s.ListErr = errors.New(listErrorMsg)

	s.Storage, err = gorm.NewStorage(
		gorm.WithInMemory(),
		gorm.WithMaxOpenConns(1),
		gorm.WithAdditionalAutoMigration(
			evaluation.EvaluationResult{},
			evaluation.Compliance{},
			collection.CollectionModule{},
			common.Evidence{},
			common.Error{}),
		// For storage debugging, set to `logger.Info`
		gorm.WithLogger(logger.Default.LogMode(logger.Silent)))
	if err != nil {
		t.Errorf("Could not initialize in-memory db: %v", err)
	}

	for _, f := range funcs {
		f(s.Storage)
	}

	return s
}

// StorageWithListError can be used to introduce various errors (without Create) in a storage operation during
// unit testing.
type StorageWithListError struct {
	persistence.Storage

	ListErr error
}

func (s *StorageWithListError) List(_ interface{}, _ string, _ bool, _ int, _ int, _ ...interface{}) error {
	return s.ListErr
}

// NewInMemoryStorageWithSaveError uses the Clouditor in-memory package to create a new in-memory storage that can be used for unit
// testing. The funcs varargs can be used to immediately execute storage operations on it.
func NewInMemoryStorageWithSaveError(t *testing.T, saveErrorMsg string, funcs ...func(s persistence.Storage)) persistence.Storage {
	var err error
	s := &StorageWithSaveError{}
	s.SaveErr = errors.New(saveErrorMsg)

	s.Storage, err = gorm.NewStorage(
		gorm.WithInMemory(),
		gorm.WithMaxOpenConns(1),
		gorm.WithAdditionalAutoMigration(
			evaluation.EvaluationResult{},
			evaluation.Compliance{},
			collection.CollectionModule{},
			common.Evidence{},
			common.Error{}),
		// For storage debugging, set to `logger.Info`
		gorm.WithLogger(logger.Default.LogMode(logger.Silent)))
	if err != nil {
		t.Errorf("Could not initialize in-memory db: %v", err)
	}

	for _, f := range funcs {
		f(s.Storage)
	}

	return s
}

// StorageWithSaveError can be used to introduce `save` error in a storage operation during unit testing.
type StorageWithSaveError struct {
	persistence.Storage

	SaveErr error
}

func (s *StorageWithSaveError) Save(_ interface{}, _ ...interface{}) error {
	return s.SaveErr
}
