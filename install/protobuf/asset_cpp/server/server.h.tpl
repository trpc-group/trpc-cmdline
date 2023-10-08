{{- $namespaces := (splitList "." .PackageName) -}}
{{- $reverseNamespaces := (splitList "." .PackageName | reverse) -}}
#pragma once

#include "trpc/common/trpc_app.h"

{{ range $val := $namespaces -}}
namespace {{ $val }} {
{{ end }}
class {{ basenamewithoutext .FilePath | camelcase }}Server : public ::trpc::TrpcApp {
 public:
  int Initialize() override;

  void Destroy() override;
};

{{ range $val := $reverseNamespaces -}}
}  // namespace {{ $val }}
{{ end -}}
