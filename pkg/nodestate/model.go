package nodestate

type NodeStateModel struct {
	startingPods  []string
	runningPods   []string
	unhealthyPods []string
}

type PodStateModel struct {
	name      string
	namespace string
}
