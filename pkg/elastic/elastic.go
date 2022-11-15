package elastic

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type Elastic struct {
	User     string
	Password string
	Address  string
}

func (c *Elastic) CreateDatabase(ctx context.Context, indexName string) (err error) {

	cfg := elasticsearch.Config{
		Addresses: []string{c.Address},
		Username:  c.User,
		Password:  c.Password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Can't create elastic client %v", err))
	}

	respExists, err := esapi.IndicesExistsRequest{Index: []string{indexName}}.Do(ctx, es)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Can't get index  %v", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			klog.Errorf("Failed to close DB connection")
		}
	}(respExists.Body)
	if respExists.StatusCode == http.StatusNotFound {
		respCreate, err := esapi.IndicesCreateRequest{
			Index: indexName,
		}.Do(ctx, es)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Can't create index %v", err))
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				klog.Errorf("Failed to close DB connection")
			}
		}(respCreate.Body)
		if respCreate.StatusCode != http.StatusOK {
			return status.Error(codes.Internal, respCreate.String())
		}
	} else if respExists.StatusCode == http.StatusOK {
		return status.Error(codes.AlreadyExists, "Index already exists")
	}

	return nil
}

func (c *Elastic) DeleteDatabase(ctx context.Context, indexName string) (err error) {
	klog.Info("Delete index ", indexName)
	cfg := elasticsearch.Config{
		Addresses: []string{c.Address},
		Username:  c.User,
		Password:  c.Password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	respExists, err := esapi.IndicesExistsRequest{Index: []string{indexName}}.Do(ctx, es)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			klog.Errorf("Failed to close DB connection")
		}
	}(respExists.Body)
	if respExists.StatusCode == http.StatusOK {
		respDelete, err := esapi.IndicesDeleteRequest{Index: []string{indexName}}.Do(ctx, es)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Can't delete index %v", err))
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				klog.Errorf("Failed to close DB connection")
			}
		}(respDelete.Body)
		if respDelete.StatusCode != http.StatusOK {
			return status.Error(codes.Internal, respDelete.String())
		}
		return nil
	}
	return status.Error(codes.NotFound, fmt.Sprintf("Index %s not found", indexName))
}
