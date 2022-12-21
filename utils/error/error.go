package error

type NoteError struct {
	Err        error
	IsNoNoter  bool
	IsNotNoter bool
}

func (e *NoteError) Unwrap() error  { return e.Err }
func (e *NoteError) Error() string  { return "error when note check: " + e.Err.Error() }
func (e *NoteError) NotNoter() bool { return e.IsNotNoter }
func (e *NoteError) NoNoter() bool  { return e.IsNoNoter }
