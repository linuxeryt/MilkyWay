// job模块注册中心

package register

import "milkyway/agent"

var ModuleMapOfJob = map[string]agent.IJob{} // 所有job模块都会注册到 ModuleMapOfJob 中

func Register(jobName string, job agent.IJob) {
	ModuleMapOfJob[jobName] = job
}