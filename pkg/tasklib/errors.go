package tasklib

import "fmt"

// ResourceConnectionError is a specific error type that indicates a task failed
// because it could not connect to or communicate with an external resource.
// When a task returns this error, it triggers a Resource-Bound Alarm.
type ResourceConnectionError struct {
	ResourceName  string
	OriginalError error
}

func (e *ResourceConnectionError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("failed to connect to resource '%s': %v", e.ResourceName, e.OriginalError)
	}
	return fmt.Sprintf("failed to connect to resource '%s'", e.ResourceName)
}

func (e *ResourceConnectionError) Unwrap() error {
	return e.OriginalError
}

// NewResourceConnectionError creates a new ResourceConnectionError.
func NewResourceConnectionError(resourceName string, err error) error {
	return &ResourceConnectionError{
		ResourceName:  resourceName,
		OriginalError: err,
	}
}
