package bendis

import (
	"fmt"
	"regexp"
	"runtime"
	"time"
)

func (b *Bendis) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	pc, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(pc)
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	b.InfoLog.Println(fmt.Sprintf("Load Time: %s took %s", name, elapsed))
}
