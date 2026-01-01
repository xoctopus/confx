package types

type LivenessChecker interface {
	LivenessCheck() map[string]string
}
