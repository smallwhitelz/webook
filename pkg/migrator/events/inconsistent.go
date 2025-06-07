package events

type InconsistentEvent struct {
	ID int64

	// 用什么来修，取值为SRC，意味着以源表为准，取值为DST，以目标表为准
	Direction string

	// 有些时候，一些观测，或者一些第三方，需要知道，是什么引起的不一致
	// 因为需要DEBUG
	// 这个是可选的
	Type string
}

const (
	// InconsistentEventTypeTargetMissing 校验的目标数据，缺了这一条
	InconsistentEventTypeTargetMissing = "target_missing"

	// InconsistentEventTypeNeq 不相等
	InconsistentEventTypeNeq = "neq"

	// InconsistentEventTypeBaseMissing 校验的源表数据，缺了这一条
	InconsistentEventTypeBaseMissing = "base_missing"
)
