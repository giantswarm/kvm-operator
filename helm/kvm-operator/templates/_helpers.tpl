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
helm.sh/chart: {{ include "kvm-operator.chart" . }}
{{ include "kvm-operator.selectorLabels" . }}
app.giantswarm.io/branch: {{ .Values.project.branch }}
app.giantswarm.io/commit: {{ .Values.project.commit }}
app.kubernetes.io/name: {{ include "kvm-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- include "kvm-operator.name" . }}.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "kvm-operator.selectorLabels" -}}
app: {{ include "kvm-operator.name" . }}
version: {{ .Chart.Version }}
{{- end -}}
