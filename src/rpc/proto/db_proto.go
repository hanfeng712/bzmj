package proto

const (
	Ok      = 0
	NoExist = 404
)

type DBQuery struct {
	Table string
	Key   string
}

type DBQueryResult struct {
	Code  uint32
	Value []byte
}

type DBDel struct {
	Table string
	Key   string
}

type DBDelResult struct {
	Code uint32
}

type DBWrite struct {
	Table string
	Key   string
	Value []byte
}

type DBWriteResult struct {
	Code uint32
}
