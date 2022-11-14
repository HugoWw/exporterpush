package model

/*
define barad_ck_push struct
*/

type BaradCk struct {
	Timestamp int        `json:"timestamp"`
	Namespace string     `json:"namespace"`
	Dimension Dimensions `json:"dimension"`
	Batch     []Batchs   `json:"batch"`
}

type Dimensions struct {
	AppId      string `json:"appid"`
	InstanceId string `json:"instanceid"`
	NodeId     string `json:"nodeid"`
	ProjectId  string `json:"projectid"`
}

type Batchs struct {
	Unit  string  `json:"unit"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
