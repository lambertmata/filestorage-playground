package fsp

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"
)

type Object struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type FileStorageProvider interface {
	Get(id string) (io.Reader, error)
	Put(name string, reader io.Reader) (*Object, error)
	List() []*Object
}

type ObjectIndex struct {
	Objects map[string]*Object
	mux     sync.RWMutex
}

func NewObjectIndex() *ObjectIndex {
	return &ObjectIndex{
		Objects: make(map[string]*Object),
	}
}

func (i *ObjectIndex) Set(o *Object) {
	i.mux.Lock()
	i.Objects[o.ID] = o
	i.mux.Unlock()
}

func (i *ObjectIndex) Get(ID string) (*Object, error) {
	i.mux.RLock()
	o, ok := i.Objects[ID]
	if !ok {
		return nil, errors.New("object not found")
	}
	i.mux.RUnlock()
	return o, nil
}

func (i *ObjectIndex) All() []*Object {
	i.mux.RLock()
	var objects []*Object
	for _, object := range i.Objects {
		objects = append(objects, object)
	}
	i.mux.RUnlock()
	return objects
}

type LocalStorage struct {
	// Path the local storage path
	Path  string
	Index *ObjectIndex
}

func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{
		Path:  path,
		Index: NewObjectIndex(),
	}
}

func (s *LocalStorage) Get(id string) (io.Reader, error) {
	pathToFile := path.Join(s.Path, id)

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return nil, errors.New("file does not exist")
	}

	return os.Open(pathToFile)
}

func (s *LocalStorage) Put(name string, reader io.Reader) (*Object, error) {

	slog.Info("writing", "name", name)

	id, err := uuid.NewUUID()

	if err != nil {
		return nil, fmt.Errorf("failed to create object id %w", err)
	}

	newObject := Object{
		Name:      name,
		ID:        id.String(),
		CreatedAt: time.Now(),
	}

	pathToFile := path.Join(s.Path, id.String())

	newFileWriter, err := os.Create(pathToFile)

	defer func() {
		newFileWriter.Close()
		s.Index.Set(&newObject)
	}()

	if err != nil {
		return nil, fmt.Errorf("failed to create file %w", err)
	}

	_, err = io.Copy(newFileWriter, reader)

	if err != nil {
		return nil, fmt.Errorf("failed to copy file %w", err)
	}

	return &newObject, nil
}

func (s *LocalStorage) List() []*Object {
	return s.Index.All()
}
