package config


type GlobalConfig struct {
	EtcdEndPoints	[]string
	AgentID			string
}


var AgentConfig   = new(GlobalConfig)