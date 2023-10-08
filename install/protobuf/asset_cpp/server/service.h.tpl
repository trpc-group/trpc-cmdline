{{- $namespaces := (splitList "." .PackageName) -}}
{{- $reverseNamespaces := (splitList "." .PackageName | reverse) -}}
{{- $pkgNamespace := (replace .PackageName "." "::") -}}
#pragma once

#include "{{ trimright ".proto" .RelatvieFilePath }}.trpc.pb.h"

{{ range $val := $namespaces -}}
namespace {{ $val }} {
{{ end }}
{{- range $svc := .Services }}
class {{ $svc.Name }}ServiceImpl : public ::{{ $pkgNamespace }}::{{ $svc.Name }} {
 public:
  {{- range $method := .RPC -}}
  {{- $rpcReqType := (replace $method.RequestType "." "::") -}}
  {{- $rpcRspType := (replace $method.ResponseType "." "::") -}}
  {{ if and $method.ClientStreaming $method.ServerStreaming }}
  ::trpc::Status {{ $method.Name }}(const ::trpc::ServerContextPtr& context, const ::trpc::stream::StreamReader<::{{ $rpcReqType }}>& reader, ::trpc::stream::StreamWriter<::{{ $rpcRspType }}>* writer) override;
  {{- else if $method.ClientStreaming }}
  ::trpc::Status {{ $method.Name }}(const ::trpc::ServerContextPtr& context, const ::trpc::stream::StreamReader<::{{ $rpcReqType }}>& reader, ::{{ $rpcRspType }}* response) override;
  {{- else if $method.ServerStreaming }}
  ::trpc::Status {{ $method.Name }}(const ::trpc::ServerContextPtr& context, const ::{{ $rpcReqType }}& request, ::trpc::stream::StreamWriter<::{{ $rpcRspType }}>* writer) override;
  {{- else }}
  ::trpc::Status {{ $method.Name }}(::trpc::ServerContextPtr context, const ::{{ $rpcReqType }}* request, ::{{ $rpcRspType }}* reply) override;
  {{- end -}}
  {{ end }}
};
{{ end }}
{{ range $val := $reverseNamespaces -}}
}  // namespace {{ $val }}
{{ end -}}
