{{/* ---------------------------------------------------------------------------
Helper templates for chart names and labels
--------------------------------------------------------------------------- */}}

{{/* Returns the base chart name */}}
{{- define "wis2-ingest.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{/* Returns the full name, combining release name + chart name
     Truncated to 63 characters (Kubernetes requirement for object names) */}}
{{- define "wis2-ingest.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Common labels to apply to all objects */}}
{{- define "wis2-ingest.labels" -}}
app.kubernetes.io/name: {{ include "wis2-ingest.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: Helm
{{- end -}}