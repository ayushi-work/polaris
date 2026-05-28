package rules

import (
	"github.com/ayushi/polaris/internal/detector"
)

func All() []detector.Rule {
	return []detector.Rule{
		&OOMKilledRule{},
		&CrashLoopRule{},
		&ImagePullRule{},
		&PodPendingRule{},
		&NodePressureRule{},
	}
}
