package katia

type Error struct {
	Message string
}

var (
	ErrInteractionForbidden = Error{
		Message: "권한이 없습니다",
	}
)

func (err Error) Error() string {
	return err.Message
}
