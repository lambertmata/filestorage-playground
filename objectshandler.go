package fsp

import (
	"fmt"
	"github.com/lambertmata/churro"
	"io"
	"net/http"
)

type ObjectsHandler struct {
	Storage FileStorageProvider
}

type ObjectCreate struct {
	Name string    `validate:"required" `
	File io.Reader `json:"file"`
}

func NewObjectsHandler(storage FileStorageProvider) *ObjectsHandler {
	return &ObjectsHandler{
		Storage: storage,
	}
}

func (o *ObjectsHandler) Put(ctx *churro.ContextWithBody[ObjectCreate]) (*churro.HandlerResponse[*Object], error) {

	uploaded, err := o.Storage.Put(ctx.Body.Name, ctx.Body.File)

	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("could not put object: %v", err)
	}

	return churro.Response(uploaded, err, churro.WithWrappedData(), churro.WithJSONContentType())
}

func (o *ObjectsHandler) Get(ctx *churro.Context) (*[]byte, error) {

	reader, err := o.Storage.Get(churro.GetPathParam(ctx.Req, "id"))

	if err != nil {
		ctx.Res.WriteHeader(http.StatusNotFound)
		return nil, fmt.Errorf("could not get object: %v", err)
	}

	bytes, err := io.ReadAll(reader)

	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("could not read object: %v", err)
	}

	return &bytes, nil
}

func (o *ObjectsHandler) List(ctx *churro.Context) (*churro.HandlerResponse[[]*Object], error) {

	objects := o.Storage.List()

	if objects == nil {
		objects = make([]*Object, 0)
	}

	return churro.Response(objects, nil, churro.WithWrappedData(), churro.WithJSONContentType())
}
