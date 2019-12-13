package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/k8s"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)


const (
	defaultVstate						  = 1
		)

func TestVps(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Provisioning", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// This test performs VolumePlacementStrategy's replica affinity  of application
// volume
var _ = Describe("{ReplicaAffinity}", func() {
	It("has to schedule app and verify the volume replica affinity", func() {

		var vpsSpec string
		vpsRules := GetVpsRules(1)

		for vkey, vrule := range vpsRules {
			var contexts []*scheduler.Context
			var volNodes map[string]map[string][]string

			var lblData []labelDict
			var setLabels int
			Step("get nodes and set labels: "+vkey, func() {
				lblData,setLabels = getTestLabels(vrule.GetLabels)
				RemoveNodeLabels(lblData)
				if setLabels == 1 {
					lblnode := SetNodeLabels(lblData)
					logrus.Debug("Nodes containing label", lblnode)
					Expect(lblnode).NotTo(BeEmpty())
					volNodes = pvcNodeMap(vrule.GetPvcNodeLabels, lblnode)
				}
			})

			Step("rules of volume placement: "+vkey, func() {
				vpsSpec = getVpsSpec(vrule.GetSpec)
			})

			Step("launch application with new vps specs :"+vkey, func() {
				applyVpsSpec(vpsSpec)
				logrus.Debugf("Spec Dir to rescan: %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)

				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaaffinity-%d", i),vrule.GetScStrategyMap(),defaultVstate )...)
				}
			})

			Step("validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate, ctx, volNodes)
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lblData)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})


// This test performs VolumePlacementStrategy's volume affinity  of application
// volume
var _ = Describe("{VolumeAffinity}", func() {
	It("has to schedule app and verify the volume replica affinity", func() {

		var vpsSpec string
		vpsRules := GetVpsRules(2)

		for vkey, vrule := range vpsRules {
			var contexts []*scheduler.Context
			var volNodes map[string]map[string][]string

			var lblData []labelDict
			var setLabels int
			Step("get nodes and set labels: "+vkey, func() {
				lblData,setLabels = getTestLabels(vrule.GetLabels)
				RemoveNodeLabels(lblData)
				if setLabels == 1 {
					lblnode := SetNodeLabels(lblData)
					logrus.Debug("Nodes containing label", lblnode)
					Expect(lblnode).NotTo(BeEmpty())
					volNodes = pvcNodeMap(vrule.GetPvcNodeLabels, lblnode)
				}
			})

			Step("rules of volume placement: "+vkey, func() {
				vpsSpec = getVpsSpec(vrule.GetSpec)
			})

			Step("launch application with new vps specs :"+vkey, func() {
				applyVpsSpec(vpsSpec)
				logrus.Debugf("Spec Dir to rescan: %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)

				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("volumeaffinity-%d", i),vrule.GetScStrategyMap() , defaultVstate)...)
				}
			})

			Step("validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate, ctx, volNodes)
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lblData)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})


// This test performs VolumePlacementStrategy's replica & volume affinity  of application
// volume
var _ = Describe("{ReplicaVolumeAffinity}", func() {
	It("has to schedule app and verify the volume replica affinity", func() {

		var vpsSpec string
		vpsRules := GetVpsRules(3)

		for vkey, vrule := range vpsRules {
			var contexts []*scheduler.Context
			var volNodes map[string]map[string][]string

			var lblData []labelDict
			var setLabels int
			Step("get nodes and set labels: "+vkey, func() {
				lblData,setLabels = getTestLabels(vrule.GetLabels)
				RemoveNodeLabels(lblData)
				if setLabels == 1 {
					lblnode := SetNodeLabels(lblData)
					logrus.Debug("Nodes containing label", lblnode)
					Expect(lblnode).NotTo(BeEmpty())
					volNodes = pvcNodeMap(vrule.GetPvcNodeLabels, lblnode)
				}
			})

			Step("rules of volume placement: "+vkey, func() {
				vpsSpec = getVpsSpec(vrule.GetSpec)
			})

			Step("launch application with new vps specs :"+vkey, func() {
				applyVpsSpec(vpsSpec)
				logrus.Debugf("Spec Dir to rescan: %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)

				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaivolumeaffinity-%d", i),vrule.GetScStrategyMap(),defaultVstate )...)
				}
			})

			Step("validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate, ctx, volNodes)
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lblData)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})


// This test performs VolumePlacementStrategy's replica & volume affinity
// with app scale Up  & Down of application
var _ = Describe("{ReplicaVolumeAffinityScale}", func() {
	It("has to schedule app and verify the volume replica affinity", func() {

		var vpsSpec string
		vpsRules := GetVpsRules(4)

		for vkey, vrule := range vpsRules {
			var contexts []*scheduler.Context
			var volNodes map[string]map[string][]string

			var lblData []labelDict
			var setLabels int
			Step("get nodes and set labels: "+vkey, func() {
				lblData,setLabels = getTestLabels(vrule.GetLabels)
				RemoveNodeLabels(lblData)
				if setLabels == 1 {
					lblnode := SetNodeLabels(lblData)
					logrus.Debug("Nodes containing label", lblnode)
					Expect(lblnode).NotTo(BeEmpty())
					volNodes = pvcNodeMap(vrule.GetPvcNodeLabels, lblnode)
				}
			})

			Step("rules of volume placement: "+vkey, func() {
				vpsSpec = getVpsSpec(vrule.GetSpec)
			})

			Step("launch application with new vps specs :"+vkey, func() {
				applyVpsSpec(vpsSpec)
				logrus.Debugf("Spec Dir to rescan: %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)

				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaivolumeaffinityscale-%d", i),vrule.GetScStrategyMap(),defaultVstate )...)
				}
			})

			Step("Scale up app", func() {
				for _, ctx := range contexts {
					Step(fmt.Sprintf("scale up app: %s by 1, Number of workernodes:%d ", ctx.App.Key, len(node.GetWorkerNodes())), func() {
						applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
						Expect(err).NotTo(HaveOccurred())
						for name, scale := range applicationScaleUpMap {
							applicationScaleUpMap[name] = scale + 1
						}
						err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
						Expect(err).NotTo(HaveOccurred())
					})

					Step("Giving few seconds for scaled up applications to stabilize", func() {
						time.Sleep(60 * time.Second)
					})

					ValidateContext(ctx,defaultVstate)
					Step("validate volumes and replica affinity: "+vkey, func() {
						for _, ctx := range contexts {
							ValidateVpsRules(vrule.Validate, ctx, volNodes)
						}
					})
					Step(fmt.Sprintf("scale down app %s by 1", ctx.App.Key), func() {
						applicationScaleDownMap, err := Inst().S.GetScaleFactorMap(ctx)
						Expect(err).NotTo(HaveOccurred())
						for name, scale := range applicationScaleDownMap {
							applicationScaleDownMap[name] = scale - 1
						}
						err = Inst().S.ScaleApplication(ctx, applicationScaleDownMap)
						Expect(err).NotTo(HaveOccurred())
					})

					Step("Giving few seconds for scaled down applications to stabilize", func() {
						time.Sleep(60 * time.Second)
					})

					ValidateContext(ctx,defaultVstate)
					Step("validate volumes and replica affinity: "+vkey, func() {
						for _, ctx := range contexts {
							ValidateVpsRules(vrule.Validate, ctx, volNodes)
						}
					})

				}

			})

			Step("validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate, ctx, volNodes)
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lblData)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})


// This test performs VolumePlacementStrategy's replica & volume affinity  of application
// with volumes pending state
var _ = Describe("{ReplicaVolumeAffinityPending}", func() {
	It("has to schedule app and verify the volume replica affinity", func() {

		var vpsSpec string
		vpsRules := GetVpsRules(5)

		for vkey, vrule := range vpsRules {
			var contexts []*scheduler.Context
			var volNodes map[string]map[string][]string

			var lblData []labelDict
			var setLabels int
			Step("get nodes and set labels: "+vkey, func() {
				lblData,setLabels = getTestLabels(vrule.GetLabels)
				RemoveNodeLabels(lblData)
				if setLabels == 1 {
					lblnode := SetNodeLabels(lblData)
					logrus.Debug("Nodes containing label", lblnode)
					Expect(lblnode).NotTo(BeEmpty())
					volNodes = pvcNodeMap(vrule.GetPvcNodeLabels, lblnode)
				}
			})

			Step("rules of volume placement: "+vkey, func() {
				vpsSpec = getVpsSpec(vrule.GetSpec)
			})

			Step("launch application with new vps specs :"+vkey, func() {
				applyVpsSpec(vpsSpec)
				logrus.Debugf("Spec Dir to rescan: %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)

				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaivolumepending-%d", i),vrule.GetScStrategyMap(), 0 )...)
				}
			})

			Step("validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate, ctx, volNodes)
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lblData)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})



var _ = Describe("{DefaultRepVolAffinity}", func() {
	It("has to schedule app and verify the volume replica affinity", func() {

		var vpsSpec string
		vpsRules := GetVpsRules(6)

		for vkey, vrule := range vpsRules {
			var contexts []*scheduler.Context
			var volNodes map[string]map[string][]string

			var lblData []labelDict
			var setLabels int
			Step("get nodes and set labels: "+vkey, func() {
				lblData,setLabels = getTestLabels(vrule.GetLabels)
				RemoveNodeLabels(lblData)
				if setLabels == 1 {
					lblnode := SetNodeLabels(lblData)
					logrus.Debug("Nodes containing label", lblnode)
					Expect(lblnode).NotTo(BeEmpty())
					volNodes = pvcNodeMap(vrule.GetPvcNodeLabels, lblnode)
				}
			})

			Step("rules of volume placement: "+vkey, func() {
				vpsSpec = getVpsSpec(vrule.GetSpec)
			})

			Step("launch application with new vps specs :"+vkey, func() {
				applyVpsSpec(vpsSpec)
				logrus.Debugf("Spec Dir to rescan: %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)

				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaaffinity-%d", i),vrule.GetScStrategyMap(),defaultVstate )...)
				}
			})

			Step("validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate, ctx, volNodes)
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lblData)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})







//-- Common function
//ValidateVpsRules checks applied vps rules on the app
func ValidateVpsRules(f func([]*volume.Volume, map[string]map[string][]string), ctx *scheduler.Context, volscheck map[string]map[string][]string) {
	var err error
	var appVolumes []*volume.Volume
	appVolumes, err = Inst().S.GetVolumes(ctx)
	Expect(err).NotTo(HaveOccurred())
	Expect(appVolumes).NotTo(BeEmpty())

	f(appVolumes, volscheck)

}

func getTestLabels(f func() ([]labelDict,int)) ([]labelDict,int) {
	return f()
}

//pvcNodeMap  The nodes on which this pvc is expected to have replica
func pvcNodeMap(f func(map[string][]string) map[string]map[string][]string, val map[string][]string) map[string]map[string][]string {

	return f(val)
}

//SetNodeLabels set the provided labels on the portworx worker nodes
func SetNodeLabels(labels []labelDict) map[string][]string {

	lblnodes := map[string][]string{}
	workerNodes := node.GetWorkerNodes()
	workerCnt := len(workerNodes)
	nodes2lbl := len(labels)

	if workerCnt < nodes2lbl {
		fmt.Printf("Required(%v) number of worker nodes(%v) not available", nodes2lbl, workerCnt)
		return lblnodes
	}

	// Get nodes
	for key, nlbl := range labels {
		//TODO: Randomize node selection
		n := workerNodes[key]
		for lkey, lval := range nlbl {
			if err := k8s.Instance().AddLabelOnNode(n.Name, lkey, lval.(string)); err != nil {
				logrus.Errorf("Failed to set node label %v: %v Err: %v", lkey, nlbl, err)
				return lblnodes
			}
			lblnodes[lkey+lval.(string)] = append(lblnodes[lkey+lval.(string)], n.Name)
		}

	}


	//for leftover nodes, labels for zone and region will be 'default'

	zonelbl := "failure-domain.beta.kubernetes.io/zonedefault"
	regionlbl := "failure-domain.beta.kubernetes.io/regiondefault"
	if workerCnt > nodes2lbl {
		for i:=(workerCnt -1); i >=nodes2lbl; i-- {
			n := workerNodes[i]
			lblnodes[zonelbl] = append(lblnodes[zonelbl], n.Name)
			lblnodes[regionlbl] = append(lblnodes[regionlbl], n.Name)
		}
	}

	//TODO: Return node list with the labels attached
	return lblnodes
}

// RemoveNodeLabels  remove the case specific lables from all nodes
func RemoveNodeLabels(labels []labelDict) {

	workerNodes := node.GetWorkerNodes()

	// Get nodes
	for _, n := range workerNodes {
		for _, nlbl := range labels {
			for lkey, lval := range nlbl {
				if err := k8s.Instance().RemoveLabelOnNode(n.Name, lkey); err != nil {
					logrus.Errorf("Failed to remove node label %v=%v: %v Err: %v", lkey, lval, nlbl, err)
					//return lblnodes
				}
			}

		}
	}
}

func getVpsSpec(f func() string) string {
	return f()
}

func applyVpsSpec(vpsSpec string) error {
	logrus.Debugf("vpsSpec:%v, ---SpecDir:%v--- App: %v ", vpsSpec, Inst().SpecDir, Inst().AppList[0])

	appDir := Inst().AppList[0]
	f, err := os.Create(Inst().SpecDir + "/" + appDir + "/vps.yaml")
	if err != nil {
		logrus.Errorf("Failed to create VPS spec: %v ", Inst().SpecDir+"/"+appDir+"/vps.yaml")
		return err
	}
	defer f.Close()

	nsize, err := f.WriteString(vpsSpec)
	if err != nil {
		logrus.Errorf("Failed to write VPS spec: %v ", Inst().SpecDir+"/"+appDir+"/vps.yaml")
		return err
	}
	f.Sync()
	logrus.Debugf("Created VPS spec: %v size: %v", Inst().SpecDir+"/"+appDir+"/vps.yaml", nsize)
	return err
}

func cleanVps() {
	logrus.Infof("Cleanup test case context")
}

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
