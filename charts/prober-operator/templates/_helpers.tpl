{{/*
Expand the name of the chart.
*/}}
{{- define "prober-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "prober-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "prober-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "prober-operator.labels" -}}
helm.sh/chart: {{ include "prober-operator.chart" . }}
{{ include "prober-operator.selectorLabels" . }}
app.kubernetes.io/component: manager
app.kubernetes.io/instance: system
control-plane: controller-manager
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "prober-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "prober-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "prober-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "prober-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Include the labels from values.yaml or use the default labels */}}
{{- define "mychart.labels" -}}
{{- default .Values.blackboxExporter.labels . | toYaml | nindent 2 }}
{{- end }}

{{/*
Release namespace
*/}}
{{- define "prober-operator.namespace" -}}
{{- .Release.Namespace }}
{{- end }}