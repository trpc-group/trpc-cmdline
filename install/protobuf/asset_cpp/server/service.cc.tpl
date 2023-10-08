{{- $namespaces := (splitList "." .PackageName) -}}
{{- $reverseNamespaces := (splitList "." .PackageName | reverse) -}}
{{- $pkgNamespace := (replace .PackageName "." "::") -}}
#include "server/service.h"

#include "trpc/log/trpc_log.h"

{{ range $val := $namespaces -}}
namespace {{ $val }} {
{{ end }}
{{- range $svc := .Services }}
{{- range $method := .RPC -}}
{{- $rpcReqType := (replace $method.RequestType "." "::") -}}
{{- $rpcRspType := (replace $method.ResponseType "." "::") -}}
{{ if and $method.ClientStreaming $method.ServerStreaming }}
::trpc::Status {{ $svc.Name }}ServiceImpl::{{ $method.Name }}(const ::trpc::ServerContextPtr& context, const ::trpc::stream::StreamReader<::{{ $rpcReqType }}>& reader, ::trpc::stream::StreamWriter<::{{ $rpcRspType }}>* writer) {
  // Please refer to examples/features/trpc_stream in trpc-cpp project
  return ::trpc::kSuccStatus;
}
{{ else if $method.ClientStreaming }}
::trpc::Status {{ $svc.Name }}ServiceImpl::{{ $method.Name }}(const ::trpc::ServerContextPtr& context, const ::trpc::stream::StreamReader<::{{ $rpcReqType }}>& reader, ::{{ $rpcRspType }}* response) {
  // Please refer to examples/features/trpc_stream in trpc-cpp project
  return ::trpc::kSuccStatus;
}
{{ else if $method.ServerStreaming }}
::trpc::Status {{ $svc.Name }}ServiceImpl::{{ $method.Name }}(const ::trpc::ServerContextPtr& context, const ::{{ $rpcReqType }}& request, ::trpc::stream::StreamWriter<::{{ $rpcRspType }}>* writer) {
  // Please refer to examples/features/trpc_stream in trpc-cpp project
  return ::trpc::kSuccStatus;
}
{{ else }}
::trpc::Status {{ $svc.Name }}ServiceImpl::{{ $method.Name }}(::trpc::ServerContextPtr context, const ::{{ $rpcReqType }}* request, ::{{ $rpcRspType }}* reply) {
  // Implement business logic here
  TRPC_FMT_INFO("got req");
  return ::trpc::kSuccStatus;
}
{{ end }}
{{- end -}}
{{ end }}
{{ range $val := $reverseNamespaces -}}
}  // namespace {{ $val }}
{{ end -}}
