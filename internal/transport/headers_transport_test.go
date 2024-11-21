package transport

import (
	"context"
	"net/http"
	"testing"

	"github.com/datolabs-io/go-backstage/v3"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestHeadersTransport_HeadersAdded(t *testing.T) {
	const baseURL = "http://localhost:7007"

	headers := map[string]string{
		"test-header-1": "test-value-1",
		"test-header-2": "test-value-2",
	}

	defer gock.Off()
	gock.New(baseURL).
		MatchHeaders(headers).
		Reply(http.StatusOK)

	client, err := backstage.NewClient(baseURL, "default", &http.Client{
		Transport: &HeadersTransport{
			Headers: headers,
		},
	})

	assert.NoErrorf(t, err, "NewClient should not return an error")
	_, _, err = client.Catalog.Entities.List(context.Background(), &backstage.ListEntityOptions{})

	assert.NoErrorf(t, err, "ListEntities should not return an error")
}
