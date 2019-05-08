package agent


type IJob interface {

	// job模块运行入口
	Run(param interface{})

	Return()	// job模块运行完成结果出口
}
