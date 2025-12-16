{{/*
Expand the name of the chart.
*/}}
{{- define "aether-defense.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "aether-defense.fullname" -}}
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
{{- define "aether-defense.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "aether-defense.labels" -}}
helm.sh/chart: {{ include "aether-defense.chart" . }}
{{ include "aether-defense.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "aether-defense.selectorLabels" -}}
app.kubernetes.io/name: {{ include "aether-defense.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "aether-defense.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "aether-defense.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get namespace
*/}}
{{- define "aether-defense.namespace" -}}
{{- if .Values.global.namespace }}
{{- .Values.global.namespace }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Get service account name for a service
*/}}
{{- define "aether-defense.serviceAccountNameForService" -}}
{{- printf "%s-%s-sa" (include "aether-defense.fullname" .) .serviceName }}
{{- end }}

{{/*
Convert service name to lowercase for Kubernetes resource names
*/}}
{{- define "aether-defense.serviceNameLower" -}}
{{- . | lower }}
{{- end }}
