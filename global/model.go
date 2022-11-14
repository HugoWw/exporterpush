package global

// MetricPoint is struct convert to timeSeries
type MetricPoint struct {
	Metric   string            `json:"metric"` // 指标名称
	LabelMap map[string]string `json:"label"`  // 数据标签
	Time     int64             `json:"time"`   // 时间戳，单位是秒
	Value    float64           `json:"value"`  // 内部字段，最终转换之后的float64数值
}
