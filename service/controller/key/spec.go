package key

type LabelsGetter interface {
	GetLabels() map[string]string
}
