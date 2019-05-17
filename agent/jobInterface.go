package agent


type IJob interface {

	// job模块运行入口
	// jobId: job id
	// param: job模块需要的相关参数
	// resultChan: 结果发送通道，job模块运行完成后将结果发送至该通道
	Run(jobID string, param map[string]interface{}, resultChan chan map[string]interface{})
	Return(resultChan chan map[string]interface{})	// job模块运行完成结果出口
}
