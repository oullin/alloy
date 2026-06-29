package workflow

import cworkflow "alloy.dev/api/contracts/workflow"

// TransitionBlocker carries a reason a guard rejected a transition.
type TransitionBlocker = cworkflow.TransitionBlocker

type TransitionBlockerList struct {
	blockers []TransitionBlocker
}

func (l *TransitionBlockerList) Add(blocker TransitionBlocker) {
	l.blockers = append(l.blockers, blocker)
}

func (l *TransitionBlockerList) All() []TransitionBlocker {
	return append([]TransitionBlocker(nil), l.blockers...)
}

func (l *TransitionBlockerList) Empty() bool {
	return len(l.blockers) == 0
}
