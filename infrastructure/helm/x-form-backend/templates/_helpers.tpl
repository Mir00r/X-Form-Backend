{{/*
Expand the name of the chart.
*/}}
{{- define "x-form-backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "x-form-backend.fullname" -}}
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
{{- define "x-form-backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "x-form-backend.labels" -}}
helm.sh/chart: {{ include "x-form-backend.chart" . }}
{{ include "x-form-backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "x-form-backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "x-form-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service labels for specific service
*/}}
{{- define "x-form-backend.serviceLabels" -}}
helm.sh/chart: {{ include "x-form-backend.chart" . }}
app.kubernetes.io/name: {{ include "x-form-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .serviceName }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Service selector labels for specific service
*/}}
{{- define "x-form-backend.serviceSelectorLabels" -}}
app.kubernetes.io/name: {{ include "x-form-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .serviceName }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "x-form-backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "x-form-backend.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate image name
*/}}
{{- define "x-form-backend.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry -}}
{{- $repository := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- else -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}
{{- end }}

{{/*
Generate service image name
*/}}
{{- define "x-form-backend.serviceImage" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry -}}
{{- $repository := .serviceConfig.image.repository -}}
{{- $tag := .serviceConfig.image.tag | default .Values.image.tag | default .Chart.AppVersion -}}
{{- if $registry -}}
{{- printf "%s/%s/%s:%s" $registry .Values.image.repository $repository $tag -}}
{{- else -}}
{{- printf "%s/%s:%s" .Values.image.repository $repository $tag -}}
{{- end -}}
{{- end }}

{{/*
Generate database URL
*/}}
{{- define "x-form-backend.databaseUrl" -}}
{{- if .Values.postgresql.enabled -}}
postgresql://{{ .Values.postgresql.auth.username }}:{{ .Values.postgresql.auth.password }}@{{ include "x-form-backend.fullname" . }}-postgresql:5432/{{ .Values.postgresql.auth.database }}
{{- else -}}
postgresql://{{ .Values.externalDatabase.username }}:$(DATABASE_PASSWORD)@{{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port }}/{{ .Values.externalDatabase.database }}
{{- end -}}
{{- end }}

{{/*
Generate Redis URL
*/}}
{{- define "x-form-backend.redisUrl" -}}
{{- if .Values.redis.enabled -}}
redis://{{ include "x-form-backend.fullname" . }}-redis:6379
{{- else -}}
{{- if .Values.externalRedis.password -}}
redis://:$(REDIS_PASSWORD)@{{ .Values.externalRedis.host }}:{{ .Values.externalRedis.port }}
{{- else -}}
redis://{{ .Values.externalRedis.host }}:{{ .Values.externalRedis.port }}
{{- end -}}
{{- end -}}
{{- end }}

{{/*
Common environment variables
*/}}
{{- define "x-form-backend.commonEnv" -}}
- name: DATABASE_URL
  value: {{ include "x-form-backend.databaseUrl" . | quote }}
- name: REDIS_URL
  value: {{ include "x-form-backend.redisUrl" . | quote }}
{{- if not .Values.postgresql.enabled }}
- name: DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.externalDatabase.existingSecret }}
      key: {{ .Values.externalDatabase.existingSecretPasswordKey }}
{{- end }}
{{- if not .Values.redis.enabled }}
{{- if .Values.externalRedis.password }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.externalRedis.existingSecret }}
      key: {{ .Values.externalRedis.existingSecretPasswordKey }}
{{- end }}
{{- end }}
{{- if .Values.observability.enabled }}
- name: OTEL_SERVICE_NAME
  value: {{ .serviceName | quote }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ .Values.config.observability.otelCollectorEndpoint | quote }}
- name: OTEL_SERVICE_VERSION
  value: {{ .Chart.AppVersion | quote }}
- name: OTEL_ENVIRONMENT
  value: {{ .Values.config.app.environment | quote }}
- name: PROMETHEUS_ENABLED
  value: {{ .Values.config.observability.prometheusEnabled | quote }}
{{- if .Values.config.observability.sentryDsn }}
- name: SENTRY_DSN
  valueFrom:
    secretKeyRef:
      name: {{ include "x-form-backend.fullname" . }}-sentry
      key: dsn
{{- end }}
{{- end }}
{{- range .Values.env }}
- name: {{ .name }}
  value: {{ .value | quote }}
{{- end }}
{{- end }}
