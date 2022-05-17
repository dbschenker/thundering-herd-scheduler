package nodestate

import "time"

type NodeStateModel struct {
	schedulingPods []PodNodeStateModel
	startingPods   []PodNodeStateModel
	runningPods    []PodNodeStateModel
	unhealthyPods  []PodNodeStateModel
}

type PodStateModel struct {
	name      string
	namespace string
}

type PodNodeStateModel struct {
	key    string
	expiry *time.Time
}
