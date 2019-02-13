package tests

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
)

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : Basic")
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// This test performs basic test of starting an application and destroying it (along with storage)
var _ = Describe("{SetupTeardown}", func() {
	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("setupteardown-%d", i))...)
		}

		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)
	})
	fmt.Println(len(contexts))

	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})

})

// Volume Driver Plugin is down, unavailable - and the client container should not be impacted.
var _ = Describe("{VolumeDriverDown}", func() {
	var contexts []*scheduler.Context
	It("has to schedule apps and stop volume driver on app nodes", func() {
		var err error
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("voldriverdown-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)
		Step("get nodes for all apps in test and bounce volume driver", func() {
			for _, ctx := range contexts {
				var appNodes []node.Node
				Step(fmt.Sprintf("get nodes for %s app", ctx.App.Key), func() {
					appNodes, err = Inst().S.GetNodesForApp(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appNodes).NotTo(BeEmpty())
				})

				Step(
					fmt.Sprintf("stop volume driver %s on app %s's nodes: %v",
						Inst().V.String(), ctx.App.Key, appNodes),
					func() {
						StopVolDriverAndWait(appNodes)
					})

				Step("starting volume driver", func() {
					StartVolDriverAndWait(appNodes)
				})

				Step("Giving few seconds for volume driver to stabilize", func() {
					time.Sleep(20 * time.Second)
				})
			}
		})

	})

	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})
})

// Volume Driver Plugin is down, unavailable on the nodes where the volumes are
// attached - and the client container should not be impacted.
var _ = Describe("{VolumeDriverDownAttachedNode}", func() {
	var contexts []*scheduler.Context
	It("has to schedule apps and stop volume driver on nodes where volumes are attached", func() {
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("voldriverdownattachednode-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)

		Step("get nodes for all apps in test and restart volume driver", func() {
			for _, ctx := range contexts {
				var appNodes []node.Node

				Step(fmt.Sprintf("get nodes for %s app", ctx.App.Key), func() {
					volumes, err := Inst().S.GetVolumes(ctx)
					Expect(err).NotTo(HaveOccurred())

					nodeMap := make(map[string]struct{})
					for _, v := range volumes {
						n, err := Inst().V.GetNodeForVolume(v)
						Expect(err).NotTo(HaveOccurred())

						if n == nil {
							continue
						}

						if _, exists := nodeMap[n.Name]; !exists {
							nodeMap[n.Name] = struct{}{}
							appNodes = append(appNodes, *n)
						}
					}
				})

				Step(fmt.Sprintf("stop volume driver %s on app %s's nodes: %v",
					Inst().V.String(), ctx.App.Key, appNodes), func() {
					StopVolDriverAndWait(appNodes)
				})

				Step("starting volume driver", func() {
					StartVolDriverAndWait(appNodes)
				})

				Step("Giving few seconds for volume driver to stabilize", func() {
					time.Sleep(20 * time.Second)
				})
			}
		})

	})

	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})

})

// Volume Driver Plugin has crashed - and the client container should not be impacted.
var _ = Describe("{VolumeDriverCrash}", func() {
	var contexts []*scheduler.Context

	It("has to schedule apps and crash volume driver on app nodes", func() {
		var err error
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("voldrivercrash-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)

		Step("get nodes for all apps in test and crash volume driver", func() {
			for _, ctx := range contexts {
				var appNodes []node.Node
				Step(fmt.Sprintf("get nodes for %s app", ctx.App.Key), func() {
					appNodes, err = Inst().S.GetNodesForApp(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appNodes).NotTo(BeEmpty())
				})

				Step(
					fmt.Sprintf("crash volume driver %s on app %s's nodes: %v",
						Inst().V.String(), ctx.App.Key, appNodes),
					func() {
						CrashVolDriverAndWait(appNodes)
					})
			}
		})
	})
	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})
})

// Volume driver plugin is down and the client container gets terminated.
// There is a lost unmount call in this case. When the volume driver is
// back up, we should be able to detach and delete the volume.
var _ = Describe("{VolumeDriverAppDown}", func() {
	var contexts []*scheduler.Context
	It("has to schedule apps, stop volume driver on app nodes and destroy apps", func() {
		var err error
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("voldriverappdown-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)

		Step("get nodes for all apps in test and bounce volume driver", func() {
			for _, ctx := range contexts {
				var appNodes []node.Node
				Step(fmt.Sprintf("get nodes for %s app", ctx.App.Key), func() {
					appNodes, err = Inst().S.GetNodesForApp(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appNodes).NotTo(BeEmpty())
				})

				var appVolumes []*volume.Volume
				Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
					appVolumes, err = Inst().S.GetVolumes(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appVolumes).NotTo(BeEmpty())
				})

				// avoid dup
				nodesThatCantBeDown := make(map[string]bool)
				nodesToBeDown := make([]node.Node, 0)
				Step(fmt.Sprintf("choose nodes to be down for %s app", ctx.App.Key), func() {
					for _, vol := range appVolumes {
						replicas, err := Inst().V.GetReplicaSetNodes(vol)
						Expect(err).NotTo(HaveOccurred())
						Expect(replicas).NotTo(BeEmpty())
						// at least n-1 nodes with replica need to be up
						for i := 0; i <= len(replicas)-1; i++ {
							nodesThatCantBeDown[replicas[i]] = true
						}
					}

					for _, node := range node.GetWorkerNodes() {
						if _, exists := nodesThatCantBeDown[node.Name]; !exists {
							nodesToBeDown = append(nodesToBeDown, node)
						}
					}

				})

				Step(fmt.Sprintf("stop volume driver %s on app %s's nodes: %v",
					Inst().V.String(), ctx.App.Key, nodesToBeDown), func() {
					StopVolDriverAndWait(nodesToBeDown)
				})

				Step(fmt.Sprintf("destroy app: %s", ctx.App.Key), func() {
					err = Inst().S.Destroy(ctx, nil)
					Expect(err).NotTo(HaveOccurred())

					Step("wait for few seconds for app destroy to trigger", func() {
						time.Sleep(10 * time.Second)
					})
				})

				Step("restarting volume driver", func() {
					StartVolDriverAndWait(nodesToBeDown)
				})

				Step(fmt.Sprintf("wait for destroy of app: %s", ctx.App.Key), func() {
					err = Inst().S.WaitForDestroy(ctx)
					Expect(err).NotTo(HaveOccurred())
				})

				DeleteVolumesAndWait(ctx)
			}
		})
	})
	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})
})

// This test deletes all tasks of an application and checks if app converges back to desired state
var _ = Describe("{AppTasksDown}", func() {
	var contexts []*scheduler.Context
	It("has to schedule app and delete app tasks", func() {
		var err error
		var contexts []*scheduler.Context
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("apptasksdown-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)

		Step("delete all application tasks", func() {
			for _, ctx := range contexts {
				Step(fmt.Sprintf("delete tasks for app: %s", ctx.App.Key), func() {
					err = Inst().S.DeleteTasks(ctx)
					Expect(err).NotTo(HaveOccurred())
				})

				ValidateContext(ctx)
			}
		})
	})

	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})

})

// This test scales up and down an application and checks if app has actually scaled accordingly
var _ = Describe("{AppScaleUpAndDown}", func() {
	var contexts []*scheduler.Context

	It("has to scale up and scale down the app", func() {
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApps(fmt.Sprintf("applicationscaleupdown-%d", i))...)
		}
		ValidateApps(fmt.Sprintf("validate apps for %s", CurrentGinkgoTestDescription().TestText), contexts)

		Step("Scale up and down all app", func() {
			for _, ctx := range contexts {
				Step(fmt.Sprintf("scale up app: %s by %d ", ctx.App.Key, len(node.GetWorkerNodes())), func() {
					applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
					Expect(err).NotTo(HaveOccurred())
					for name, scale := range applicationScaleUpMap {
						applicationScaleUpMap[name] = scale + int32(len(node.GetWorkerNodes()))
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
					Expect(err).NotTo(HaveOccurred())
				})

				Step("Giving few seconds for scaled up applications to stabilize", func() {
					time.Sleep(10 * time.Second)
				})

				ValidateContext(ctx)

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
					time.Sleep(10 * time.Second)
				})

				ValidateContext(ctx)
			}
		})
	})
	AfterEach(func() {
		TearDownAfterEachSpec(contexts)
	})

	JustAfterEach(func() {
		DescribeNamespaceJustAfterEachSpec(contexts)
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	CollectSupport()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
