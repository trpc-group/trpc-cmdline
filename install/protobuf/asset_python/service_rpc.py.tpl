# Tencent is pleased to support the open source community by making tRPC-Python available.
#
# Copyright (C) 2026 Tencent. All rights reserved.
#
# tRPC-Python is licensed under Apache-2.0.

"""Business logic module """

{{- $legacy := eq .PythonSDK "trpc" -}}
{{- $pkgName := .PackageName -}}
{{- $pbFile := trimright (last (splitList "." .Protofile)) .Protofile -}}
{{- $pbFile := trimright "." $pbFile -}}
{{- $pkgpath := gopkg $pkgName }}

{{- $importStream := false }}
{{- $importContext := false }}
{{- range $service := .Services }}
{{- range $rpc := $service.RPC }}
{{- if or $rpc.ClientStreaming $rpc.ServerStreaming }}
{{- $importStream = true }}
{{- end }}
{{- if and (ne $rpc.ClientStreaming true) (ne $rpc.ServerStreaming true) }}
{{- $importContext = true }}
{{- end }}
{{- end }}
{{- end }}

{{- if $importContext }}
{{- if $legacy }}
from trpc.context import TrpcContext
{{- else }}
from trpc_sdk.abc import TrpcContext
{{- end }}
{{- end }}

from {{$pkgpath}} import pb
from {{$pkgpath}} import rpc
{{- if $importStream }}
from {{$pkgpath}} import rpc_stream
{{- end }}

{{ $service := (index .Services .ServiceIndex) -}}
{{- $svrName := $service.Name }}
{{- $svrNameCamelCase := $service.Name | camelcase }}
class {{$svrNameCamelCase}}Servicer(rpc.{{$svrNameCamelCase}}Servicer):
    """Provides methods that implement functionality of {{$svrNameCamelCase}} servicer."""
    
    def __init__(self):
        pass
    {{ range $rpc := $service.RPC }}
    {{- $rpcName := .Name -}}
    {{- $rpcNameCamelCase := $rpcName | camelcase }}
    {{- $req := (last (splitList "." $rpc.RequestType)) }}
    {{- $rsp := (last (splitList "." $rpc.ResponseType)) }}

    {{- if and (ne $rpc.ClientStreaming true) (ne $rpc.ServerStreaming true) }}
    #pylint: disable=invalid-name
    async def {{$rpcName}}(self, ctx: TrpcContext, request: pb.{{$req}}) -> pb.{{$rsp}}:
        """Abstract rpc method needed to implementation"""
        raise NotImplementedError('Method not implemented!')
    {{- else }}
    {{- if $rpc.ClientStreaming }}
    #pylint: disable=invalid-name
    async def {{$rpcName}}(self, stream: rpc_stream.{{$svrNameCamelCase}}ServerStream.{{$rpcNameCamelCase}}):
        """Abstract client stream method needed to implementation"""
        raise NotImplementedError('Method not implemented!')
    {{- else }}
    #pylint: disable=invalid-name
    async def {{$rpcName}}(self, request: pb.{{$req}}, stream: rpc_stream.{{$svrNameCamelCase}}ServerStream.{{$rpcNameCamelCase}}):
        """Abstract stream server method needed to implementation"""
        raise NotImplementedError('Method not implemented!')
    {{- end }}
    {{- end }}
    {{ end }}
{{/* end */}}
