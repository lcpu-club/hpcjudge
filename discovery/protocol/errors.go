package protocol

import "fmt"

var ErrNoServiceAvailable = fmt.Errorf("no service available")
var ErrServiceAlreadyExists = fmt.Errorf("service already exists")
var ErrServiceDoesNotExist = fmt.Errorf("service does not exist")
var ErrNoServiceInformed = fmt.Errorf("no service informed")
var ErrUnknownOperation = fmt.Errorf("unknown operation")
