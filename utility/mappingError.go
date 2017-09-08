package utility

// MappingError is an error that can be returned by the content placeholder mapper
type MappingError struct {
	ContentUUID  string
	ErrorMessage string
}

// NewMappingError returns a new instance of a MappingError
func NewMappingError() *MappingError {
	return &MappingError{}
}

func (e MappingError) Error() string {
	return e.ErrorMessage
}

// WithMessage adds a message to a mapping error
func (e *MappingError) WithMessage(errorMsg string) *MappingError {
	e.ErrorMessage = errorMsg
	return e
}

// ForContent associate the mapping error to a specific piece of content
func (e *MappingError) ForContent(uuid string) *MappingError {
	e.ContentUUID = uuid
	return e
}
