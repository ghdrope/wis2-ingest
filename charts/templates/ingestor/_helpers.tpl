{{/* ---------------------------------------------------------------------------
Helper templates for chart names and labels
--------------------------------------------------------------------------- */}}

{{/* Base chart name, always "wis2-ingest" */}}
{{- define "wis2-ingest.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{/* Full deployment name: release + component */}}
{{- define "wis2-ingest.fullname" -}}
{{- printf "%s-%s" .Release.Name (default "ingestor" .Values.ingestor.name) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Component name for labels */}}
{{- define "wis2-ingest.component" -}}
{{- default "ingestor" .Values.ingestor.name -}}
{{- end -}}

{{/* Common labels applied to all objects */}}
{{- define "wis2-ingest.labels" -}}
app.kubernetes.io/app: {{ include "wis2-ingest.name" . }}
app.kubernetes.io/chart: {{ .Release.Name }}
app.kubernetes.io/component: {{ include "wis2-ingest.component" . }}
app.kubernetes.io/version: "{{ .Chart.AppVersion }}"
app.kubernetes.io/managed-by: Helm
{{- end -}}

{{/* Selector labels (used in Deployment spec.selector) */}}
{{- define "wis2-ingest.selectorLabels" -}}
app.kubernetes.io/name: {{ include "wis2-ingest.name" . }}
app.kubernetes.io/component: {{ include "wis2-ingest.component" . }}
{{- end -}}