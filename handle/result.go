package handle

import (
	"io"
	"log"
)

type Result struct {
	Ok   bool        `json:"ok"`
	Err  string      `json:"err"`
	Data interface{} `json:"data"`
}

func writeResult(w io.Writer, dataAndError ...interface{}) {
	var data interface{}
	if len(dataAndError) > 0 {
		data = dataAndError[0]
	}

	errorStr := ""
	if len(dataAndError) > 1 {
		errorStr = dataAndError[1].(string)
	}

	err := json.NewEncoder(w).Encode(Result{Data: data, Err: errorStr, Ok: len(errorStr) == 0})
	if err != nil {
		log.Printf("fail to write result: %v", err)
	}
}

func writeErrorResult(w io.Writer, errorStr string) {
	err := json.NewEncoder(w).Encode(Result{Err: errorStr, Ok: false})
	if err != nil {
		log.Printf("fail to write result: %v", err)
	}
}
