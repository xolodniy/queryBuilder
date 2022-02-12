package queryBuilder

import (
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

// Frame is short format of runtime.Frame
type Frame struct {
	Function string
	File     string
	Line     int
}

// GetFrames function for retrieve calling trace,
// can be used if you want write to logs calling trace
func (qb *QB) GetFrames() []Frame {
	maxLenght := make([]uintptr, 99)
	// skip firs 2 callers which is "runtime.Callers" and common.GetFrames
	n := runtime.Callers(2, maxLenght)

	var res []Frame
	if n > 0 {
		frames := runtime.CallersFrames(maxLenght[:n])
		for more, frameIndex := true, 0; more; frameIndex++ {

			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()

			// skip tracing external dependencies
			if !strings.Contains(frameCandidate.Function, qb.projectName) {
				break
			}
			res = append(res, Frame{
				Function: frameCandidate.Function,
				File:     frameCandidate.File,
				Line:     frameCandidate.Line,
			})
		}
	}
	return res
}

func initLogTrace(trace logrus.Fields) logrus.Fields {
	if trace == nil {
		return make(logrus.Fields)
	}
	return trace
}
