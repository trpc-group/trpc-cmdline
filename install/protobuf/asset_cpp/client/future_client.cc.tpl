{{- $pkgNamespace := (replace .PackageName "." "::") -}}
{{- $app := .AppName -}} {{- if eq (len .AppName) 0 -}} {{- $app = "appdemo" -}} {{- end -}}
{{- $server := .ServerName -}} {{- if eq (len .ServerName) 0 -}}  {{- $server = "serverdemo" -}} {{- end -}}
#include <iostream>
#include <string>

#include "gflags/gflags.h"

#include "trpc/client/make_client_context.h"
#include "trpc/client/trpc_client.h"
#include "trpc/common/runtime_manager.h"
#include "trpc/log/trpc_log.h"
#include "trpc/util/thread/latch.h"

#include "{{ trimright ".proto" .RelatvieFilePath }}.trpc.pb.h"

DEFINE_string(client_config, "trpc_cpp.yaml", "framework client config file, --client_config=trpc_cpp.yaml");
{{- range $idx, $svc := .Services -}}
{{- $idx_suffix := $idx -}}{{- if (eq $idx 0) -}}{{- $idx_suffix = "" -}}{{- end }}
DEFINE_string(service_name{{ $idx_suffix }}, "trpc.{{ $app }}.{{ $server }}.{{ $svc.Name }}", "callee service name");
{{- end }}

{{ range $idx, $svc := .Services }}
{{ range $method := .RPC }}
{{- $rpcReqType := (replace $method.RequestType "." "::") -}}
{{- $rpcRspType := (replace $method.ResponseType "." "::") -}}
int {{ $svc.Name }}Async{{ $method.Name }}(const std::shared_ptr<::{{ $pkgNamespace }}::{{ $svc.Name }}ServiceProxy>& proxy) {
  {{- if and $method.ClientStreaming $method.ServerStreaming }}
  ::trpc::ClientContextPtr client_ctx = ::trpc::MakeClientContext(proxy);
  // Please refer to examples/features/trpc_async_stream in trpc-cpp project
  return 0;
  {{- else if $method.ClientStreaming }}
  ::trpc::ClientContextPtr client_ctx = ::trpc::MakeClientContext(proxy);
  // Please refer to examples/features/trpc_async_stream in trpc-cpp project
  return 0;
  {{- else if $method.ServerStreaming }}
  ::trpc::ClientContextPtr client_ctx = ::trpc::MakeClientContext(proxy);
  // Please refer to examples/features/trpc_async_stream in trpc-cpp project
  return 0;
  {{- else }}
  ::trpc::ClientContextPtr client_ctx = ::trpc::MakeClientContext(proxy);
  ::{{ $rpcReqType }} req;
  // fill some filed of req
  bool succ = true;
  ::trpc::Latch latch(1);
  proxy->Async{{ $method.Name }}(client_ctx, req).
      Then([&latch, &succ](::trpc::Future<::{{ $rpcRspType }}>&& fut) {
        if (fut.IsReady()) {
          auto rsp = fut.GetValue0();
          // print some filed of rsp
          std::cout << "get rsp success" << std::endl;
        } else {
          auto exception = fut.GetException();
          succ = false;
          std::cerr << "get rpc error: " << exception.what() << std::endl;
        }
        latch.count_down();
        return ::trpc::MakeReadyFuture<>();
      });
  latch.wait();
  return succ ? 0 : -1;
  {{- end }}
}

{{ end -}}
{{ end }}
int Run() {
  {{- $svc0 := (index .Services 0) -}}
  {{- if and (eq (len .Services) 1) (eq (len $svc0.RPC) 1) -}}
  {{- $method := (index $svc0.RPC 0) }}
  auto proxy = ::trpc::GetTrpcClient()->GetProxy<::{{ $pkgNamespace }}::{{ $svc0.Name }}ServiceProxy>(FLAGS_service_name);
  return {{ $svc0.Name }}Async{{ $method.Name }}(proxy);
  {{- else }}
  int ret = 0;
  {{ range $idx, $svc := .Services }}
  {{- $idx_suffix := $idx -}}{{- if (eq $idx 0) -}}{{- $idx_suffix = "" -}}{{- end }}
  auto proxy{{ $idx_suffix }} = ::trpc::GetTrpcClient()->GetProxy<::{{ $pkgNamespace }}::{{ $svc.Name }}ServiceProxy>(FLAGS_service_name{{ $idx_suffix }});
  {{ range $method := .RPC -}}
  ret = {{ $svc.Name }}Async{{ $method.Name }}(proxy{{ $idx_suffix }});
  if (ret < 0) return ret;
  {{ end -}}
  {{ end }}
  return ret;
  {{- end }}
}

void ParseClientConfig(int argc, char* argv[]) {
  google::ParseCommandLineFlags(&argc, &argv, true);
  google::CommandLineFlagInfo info;
  if (GetCommandLineFlagInfo("client_config", &info) && info.is_default) {
    TRPC_FMT_ERROR("start client with config, for example:{} --client_config=/client/config/filepath", argv[0]);
    exit(-1);
  }
  {{ range $idx, $svc := .Services -}}
  {{- $idx_suffix := $idx -}}{{- if (eq $idx 0) -}}{{- $idx_suffix = "" -}}{{- end }}
  std::cout << "FLAGS_service_name{{ $idx_suffix }}: " << FLAGS_service_name{{ $idx_suffix }} << std::endl;
  {{- end }}
  std::cout << "FLAGS_client_config: " << FLAGS_client_config << std::endl;

  int ret = ::trpc::TrpcConfig::GetInstance()->Init(FLAGS_client_config);
  if (ret != 0) {
    std::cerr << "load client_config failed." << std::endl;
    exit(-1);
  }
}

int main(int argc, char* argv[]) {
  ParseClientConfig(argc, argv);
  // If the business code is running in trpc pure client mode,
  // the business code needs to be running in the `RunInTrpcRuntime` function
  // This function can be seen as a program entry point and should be called only once.
  return ::trpc::RunInTrpcRuntime([]() {
    return Run();
  });
}
