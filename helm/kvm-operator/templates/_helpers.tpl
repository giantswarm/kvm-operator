{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kvm-operator.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kvm-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "kvm-operator.labels" -}}
helm.sh/chart: {{ include "kvm-operator.chart" . | quote }}
app: {{ include "kvm-operator.name" . | quote }}
{{ include "kvm-operator.selectorLabels" . | quote }}
app.giantswarm.io/branch: {{ .Values.project.branch | quote }}
app.giantswarm.io/commit: {{ .Values.project.commit | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "kvm-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kvm-operator.name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end -}}
