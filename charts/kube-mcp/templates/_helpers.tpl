{{/*
Expand the name of the chart.
*/}}
{{- define "kube-mcp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kube-mcp.fullname" -}}
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
{{- define "kube-mcp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kube-mcp.labels" -}}
helm.sh/chart: {{ include "kube-mcp.chart" . }}
{{ include "kube-mcp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kube-mcp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-mcp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kube-mcp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kube-mcp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create image reference
*/}}
{{- define "kube-mcp.image" -}}
{{- if .Values.global.imageRegistry }}
{{- printf "%s/%s:%s" .Values.global.imageRegistry .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- else }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}
{{- end }}

{{/*
Extract HTTP port from address
*/}}
{{- define "kube-mcp.httpPort" -}}
{{- $parts := splitList ":" .Values.server.http.address }}
{{- index $parts 1 | default "8080" | int }}
{{- end }}

{{/*
Generate config.toml content
*/}}
{{- define "kube-mcp.config" -}}
[server]
transports = {{ .Values.server.transports | toJson }}
log_level = "{{ .Values.server.logLevel }}"

[server.http]
address = "{{ .Values.server.http.address }}"

[server.http.oauth]
enabled = {{ .Values.server.http.oauth.enabled }}
provider = "{{ .Values.server.http.oauth.provider }}"
{{- if .Values.server.http.oauth.issuerURL }}
issuer_url = "{{ .Values.server.http.oauth.issuerURL }}"
{{- end }}
{{- if .Values.server.http.oauth.clientID }}
client_id = "{{ .Values.server.http.oauth.clientID }}"
{{- end }}
{{- if .Values.server.http.oauth.clientSecret }}
client_secret = "{{ .Values.server.http.oauth.clientSecret }}"
{{- else if .Values.server.http.oauth.clientSecretFromSecret }}
client_secret = "{{ .Values.server.http.oauth.clientSecretFromSecret }}"
{{- end }}
{{- if .Values.server.http.oauth.scopes }}
scopes = {{ .Values.server.http.oauth.scopes | toJson }}
{{- end }}
{{- if .Values.server.http.oauth.redirectURL }}
redirect_url = "{{ .Values.server.http.oauth.redirectURL }}"
{{- end }}

[server.http.cors]
enabled = {{ .Values.server.http.cors.enabled }}
allowed_origins = {{ .Values.server.http.cors.allowedOrigins | toJson }}
allowed_methods = {{ .Values.server.http.cors.allowedMethods | toJson }}
allowed_headers = {{ .Values.server.http.cors.allowedHeaders | toJson }}


[kubernetes]
provider = "{{ .Values.kubernetes.provider }}"
{{- if .Values.kubernetes.kubeconfigPath }}
kubeconfig_path = "{{ .Values.kubernetes.kubeconfigPath }}"
{{- end }}
{{- if .Values.kubernetes.context }}
context = "{{ .Values.kubernetes.context }}"
{{- end }}
{{- if .Values.kubernetes.defaultNamespace }}
default_namespace = "{{ .Values.kubernetes.defaultNamespace }}"
{{- end }}
qps = {{ .Values.kubernetes.qps }}
burst = {{ .Values.kubernetes.burst }}
timeout = "{{ .Values.kubernetes.timeout }}"

[security]
read_only = {{ .Values.security.readOnly }}
non_destructive = {{ .Values.security.nonDestructive }}
{{- if .Values.security.deniedGVKs }}
denied_gvks = {{ .Values.security.deniedGVKs | toJson }}
{{- end }}
require_rbac = {{ .Values.security.requireRBAC }}
validate_token = {{ .Values.security.validateToken }}
rbac_cache_ttl = {{ .Values.security.rbacCacheTTL }}

[helm]
storage_driver = "{{ .Values.helm.storageDriver }}"
default_namespace = "{{ .Values.helm.defaultNamespace }}"

[kubevirt]
enabled = {{ .Values.kubevirt.enabled }}
vm_group_version = "{{ .Values.kubevirt.vmGroupVersion }}"

[kiali]
enabled = {{ .Values.kiali.enabled }}
{{- if .Values.kiali.url }}
url = "{{ .Values.kiali.url }}"
{{- end }}
{{- if .Values.kiali.token }}
token = "{{ .Values.kiali.token }}"
{{- end }}
timeout = "{{ .Values.kiali.timeout }}"

[kiali.tls]
enabled = {{ .Values.kiali.tls.enabled }}
{{- if .Values.kiali.tls.caFile }}
ca_file = "{{ .Values.kiali.tls.caFile }}"
{{- end }}
{{- if .Values.kiali.tls.certFile }}
cert_file = "{{ .Values.kiali.tls.certFile }}"
{{- end }}
{{- if .Values.kiali.tls.keyFile }}
key_file = "{{ .Values.kiali.tls.keyFile }}"
{{- end }}
insecure_skip_verify = {{ .Values.kiali.tls.insecureSkipVerify }}
{{- end }}

