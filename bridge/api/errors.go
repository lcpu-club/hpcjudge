package api

import "fmt"

var ErrPartitionNotFound = fmt.Errorf("partition not found")
var ErrPathOverflowsPartitionPath = fmt.Errorf("path overflows partition path")
var ErrFileCreationError = fmt.Errorf("file creation error")
var ErrFailedToFetch = fmt.Errorf("failed to fetch")
