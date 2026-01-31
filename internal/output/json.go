package output

import (
	"encoding/json"
	"io"

	bpfsv1 "github.com/Tjaarda1/bpfstats/api/v1"
)

type JsonOutput struct{}

func (p *JsonOutput) OutputParam(par bpfsv1.Parameter, w io.Writer) error {

	data, err := json.MarshalIndent(par, "", "    ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}
