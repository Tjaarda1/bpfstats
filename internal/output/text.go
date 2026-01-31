package output

import (
	"io"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
)

type TextOutput struct{}

func (p *TextOutput) PrintParameter(par bpfsv1.Parameter, w io.Writer) error {

	return nil
}
