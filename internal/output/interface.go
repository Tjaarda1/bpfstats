package output

import (
	"io"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
)

type ParameterOutputFunc func(bpfsv1.Parameter, io.Writer) error

func (fn ParameterOutputFunc) OutputParam(par bpfsv1.Parameter, w io.Writer) error {
	return fn(par, w)
}

type ParameterOutput interface {
	OutputParam(bpfsv1.Parameter, io.Writer) error
}

type OutputOptions struct {
	NoHeaders    bool
	ShowLabels   bool
	ColumnLabels []string

	SortBy string

	AllowMissingKeys bool
}
