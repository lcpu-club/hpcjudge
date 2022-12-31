package api

import "fmt"

var ErrPartitionNotFound = fmt.Errorf("partition not found")
var ErrPathOverflowsPartitionPath = fmt.Errorf("path overflows partition path")
var ErrFileCreationError = fmt.Errorf("file creation error")
var ErrFailedToFetch = fmt.Errorf("failed to fetch")
var ErrFailedToLookupUser = fmt.Errorf("failed to lookup user")
var ErrFailedToCreateCommandPipe = fmt.Errorf("failed to create command pipe")
var ErrFailedToStartCommand = fmt.Errorf("failed to start command")
var ErrFailedToReadFromPipe = fmt.Errorf("failed to read from pipe")
var ErrFailedToExecuteCommand = fmt.Errorf("failed to execute command")
var ErrFailedToStatFile = fmt.Errorf("failed to stat file")
var ErrFileNotFound = fmt.Errorf("file not found")
var ErrFailedToRemove = fmt.Errorf("failed to remove")
