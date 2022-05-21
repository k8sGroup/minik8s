package jobcontroller

import (
	"context"
	"encoding/json"
	"minik8s/cmd/kube-controller-manager/util"
	"minik8s/object"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/listerwatcher"
	concurrentmap "minik8s/util/map"
	"time"
)

type JobController struct {
	ls            *listerwatcher.ListerWatcher
	jobMap        *concurrentmap.ConcurrentMapTrait[string, object.VersionedGPUJob]
	jobStatusMap  *concurrentmap.ConcurrentMapTrait[string, object.VersionedJobStatus]
	apiServerBase string
	stopChannel   chan struct{}
}

func NewJobController(ctx context.Context, controllerCtx util.ControllerContext) *JobController {
	jc := &JobController{
		ls:            controllerCtx.Ls,
		stopChannel:   make(chan struct{}),
		jobMap:        concurrentmap.NewConcurrentMapTrait[string, object.VersionedGPUJob](),
		jobStatusMap:  concurrentmap.NewConcurrentMapTrait[string, object.VersionedJobStatus](),
		apiServerBase: "http://" + controllerCtx.MasterIP + ":" + controllerCtx.HttpServerPort,
	}
	if jc.apiServerBase == "" {
		klog.Fatalf("uninitialized apiserver base!\n")
	}
	return jc
}

func (jc *JobController) Run(ctx context.Context) {
	klog.Debugf("[JobController] running...\n")
	jc.register()
	<-ctx.Done()
	close(jc.stopChannel)
}

func (jc *JobController) register() {
	registerPutJob := func() {
		for {
			err := jc.ls.Watch("/registry/job/default", jc.putJob, jc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/job\n")
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}

	registerDelJob := func() {
		for {
			err := jc.ls.Watch("/registry/job/default", jc.delJob, jc.stopChannel)
			if err != nil {
				klog.Errorf("Error watching /registry/job\n")
			} else {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}

	go registerPutJob()
	go registerDelJob()
}

func (jc *JobController) putJob(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.PUT {
		return
	}
	// TODO
	job := object.GPUJob{}
	err := json.Unmarshal(res.ValueBytes, &job)
	if err != nil {
		klog.Errorf("%s\n", err.Error())
		return
	}
	object.PodTemplate{
		ObjectMeta: object.ObjectMeta{},
		Spec:       object.PodSpec{},
	}
}

func (jc *JobController) delJob(res etcdstore.WatchRes) {
	if res.ResType != etcdstore.DELETE {
		return
	}
	// TODO
}
