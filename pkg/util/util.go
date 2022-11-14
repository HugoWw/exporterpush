package util

import (
	"fmt"
	"runtime"
	"strconv"
)

// Decimal float64 Keep the number of decimal places
// value float64 type
// prec int The number of digits after the decimal point must be reserved
func Decimal(value float64, prec int) float64 {
	value, _ = strconv.ParseFloat(strconv.FormatFloat(value, 'f', prec, 64), 64)
	return value
}

/*
CatchException
:Catch exceptions: try...catch
  examples：
    defer util.CatchException(func(e interface{}) {
      log.Println(e)
    })
*/
func CatchException(handle func(e interface{})) {
	if err := recover(); err != nil {
		e := printStackTrace(err)
		handle(e)
	}
}

// Printing stack information
func printStackTrace(err interface{}) string {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	er, ok := err.(error)
	if !ok {
		er = fmt.Errorf("%v", err)
	}

	return fmt.Sprintf("error:%v,error string:%v\n", er, string(buf))
}
