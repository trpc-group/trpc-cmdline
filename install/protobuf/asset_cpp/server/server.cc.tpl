{{- $serverName := (basenamewithoutext .FilePath | camelcase) -}}
{{- $namespaces := (splitList "." .PackageName) -}}
{{- $reverseNamespaces := (splitList "." .PackageName | reverse) -}}
{{- $pkgNamespace := (replace .PackageName "." "::") -}}
#include "server/server.h"

#include <memory>
#include <string>

#include "fmt/format.h"

#include "trpc/log/trpc_log.h"

#include "server/service.h"

{{ range $val := $namespaces -}}
namespace {{ $val }} {
{{ end }}
// The initialization logic depending on framework runtime(threadmodel,transport,etc) or plugins(config,metrics,etc) should be placed here.
// Others can simply place at main function(before Main() function invoked)
int {{ $serverName }}Server::Initialize() {
  const auto& config = ::trpc::TrpcConfig::GetInstance()->GetServerConfig();

  // Set the service name, which must be the same as the value of the `/server/service/name` configuration item in the yaml file,
  // otherwise the framework cannot receive requests normally.
  {{- range $idx, $svc := .Services -}}
  {{- $idx_suffix := $idx -}}{{- if (eq $idx 0) -}}{{- $idx_suffix = "" -}}{{- end }}
  std::string service_name{{$idx_suffix}} = fmt::format("{}.{}.{}.{}", "trpc", config.app, config.server, "{{$svc.Name}}");
  ::trpc::ServicePtr my_service{{$idx_suffix}}(std::make_shared<{{$svc.Name}}ServiceImpl>());
  RegisterService(service_name{{$idx_suffix}}, my_service{{$idx_suffix}});
  TRPC_FMT_INFO("Register service: {}", service_name{{$idx_suffix}});
  {{ end }}
  return 0;
}

// If the resources initialized in the initialize function need to be destructed when the program exits, it is recommended to place them here.
void {{ $serverName }}Server::Destroy() {}

{{ range $val := $reverseNamespaces -}}
}  // namespace {{ $val }}
{{ end }}
int main(int argc, char** argv) {
  ::{{ $pkgNamespace }}::{{ $serverName }}Server {{ $serverName | lower }}_server;
  {{ $serverName | lower }}_server.Main(argc, argv);
  {{ $serverName | lower }}_server.Wait();

  return 0;
}
