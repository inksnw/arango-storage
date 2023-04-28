package main

import (
	plugin "github.com/clusterpedia-io/arango-storage-layer/storage"
)

func init() {
	plugin.RegisterStorageLayer()
}
