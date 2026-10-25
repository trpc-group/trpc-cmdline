# Tencent is pleased to support the open source community by making tRPC-Python available.
#
# Copyright (C) 2026 Tencent. All rights reserved.
#
# tRPC-Python is licensed under Apache-2.0.

"""This is test client module"""
{{- $legacy := eq .PythonSDK "trpc" }}

import os
import sys
import asyncio
{{ if $legacy }}
from trpc.context import TrpcContext
from trpc.config import config as trpc_config
from trpc.client.options import with_target
from trpc.client.options import with_protocol
from trpc.client.options import with_serialization_type
from trpc.client.options import with_network
from trpc.client.options import with_timeout
from trpc.codec.serialization import SerializationType
from trpc.log import logger
from trpc.plugin import plugin as trpc_plugin
{{- else }}
from trpc_sdk.abc import TrpcContext
from trpc_sdk.client import Options
from trpc_sdk.config import TrpcConfig
from trpc_sdk.log import logger
from trpc_sdk.serialization import SerializationType
{{- end }}

# add the stub to the path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "stub")))  # noqa

{{- $pkgName := .PackageName -}}
{{- $pbFile := trimright (last (splitList "." .Protofile)) .Protofile -}}
{{- $pbFile := trimright "." $pbFile -}}
{{- $pkgpath := gopkg $pkgName -}}
{{- $importStream := false }}
{{- range $service := .Services }}
{{- range $rpc := $service.RPC }}
{{- if or $rpc.ClientStreaming $rpc.ServerStreaming }}
{{- $importStream = true }}
{{- end }}
{{- end }}
{{- end }}


from {{$pkgpath}} import pb
from {{$pkgpath}} import rpc
{{- if $importStream }}
from {{$pkgpath}} import rpc_stream
{{- end }}

async_test = []
sync_test = []
{{- $svrName := (index .Services 0).Name -}}
{{range  $index, $service := .Services -}}
{{ $svrName := $service.Name -}}
{{- $svrNameCamelCase := $service.Name | camelcase }}
    {{ range $rpc := $service.RPC }}
        {{- $rpcName := .Name -}}
        {{- $rpcNameCamelCase := $rpcName | camelcase }}
        {{- $req := (last (splitList "." $rpc.RequestType)) }}
        {{- $rsp := (last (splitList "." $rpc.ResponseType)) }}
        {{- if and (ne $rpc.ClientStreaming true) (ne $rpc.ServerStreaming true) }}


# pylint: disable=invalid-name
def test_sync_{{$svrNameCamelCase}}_{{$rpcName}}():
    """Test {{$svrNameCamelCase}} of {{$rpcName}}"""
    proxy = rpc.{{$svrNameCamelCase}}ClientProxyImpl()
    req = pb.{{$req}}()
    {{- if $legacy }}
    options = [
        with_target('ip://127.0.0.1:{{add 8000 $index}}'),
        with_protocol('trpc'),
        with_serialization_type(SerializationType.PB),
        with_network('tcp'),
        with_timeout(1000)
    ]
    {{- else }}
    options = Options().set_target('ip://127.0.0.1:{{add 8000 $index}}').set_protocol('trpc').set_serialization_type(
        SerializationType.PB).set_network('tcp').set_timeout(1000)
    {{- end }}
    ctx = TrpcContext()
    try:
        ret = proxy.{{$rpcName}}(ctx, req, options)
        print(f"{{$rpcName}} ret is {ret}")
        assert ret is not None
    except Exception as exc:  # pylint: disable=broad-exception-caught
        logger.error(exc)
        assert 0
sync_test.append(test_sync_{{$svrNameCamelCase}}_{{$rpcName}})

# pylint: disable=invalid-name
async def test_{{$svrNameCamelCase}}_{{$rpcName}}():
    """Async test {{$svrNameCamelCase}} of {{$rpcName}}"""
    proxy = rpc.{{$svrNameCamelCase}}ClientProxyImpl()
    req = pb.{{$req}}()
    {{- if $legacy }}
    options = [
        with_target('ip://127.0.0.1:{{add 8000 $index}}'),
        with_protocol('trpc'),
        with_serialization_type(SerializationType.PB),
        with_timeout(1000),
        with_network('tcp')
    ]
    {{- else }}
    options = Options().set_target('ip://127.0.0.1:{{add 8000 $index}}').set_protocol('trpc').set_serialization_type(
        SerializationType.PB).set_network('tcp').set_timeout(1000)
    {{- end }}
    ctx = TrpcContext()
    try:
        ret = await proxy.async{{$rpcName}}(ctx, req, options)
        print(f"{{$rpcName}} ret is {ret}")
        assert ret is not None
    except Exception as exc:  # pylint: disable=broad-exception-caught
        logger.error(exc)
        assert 0
async_test.append(test_{{$svrNameCamelCase}}_{{$rpcName}}())
        {{- else }}
# pylint: disable=invalid-name
async def test_{{$svrNameCamelCase}}_{{$rpcName}}():
    """Async test stream {{$svrNameCamelCase}} of {{$rpcName}}"""
    proxy = rpc.{{$svrNameCamelCase}}ClientProxyImpl()
    {{- if $legacy }}
    options = [
        with_target('ip://127.0.0.1:{{add 8000 $index}}'),
        with_protocol('trpc'),
        with_serialization_type(SerializationType.PB),
        with_timeout(10000),
        with_network('tcp')
    ]
    {{- else }}
    options = Options().set_target('ip://127.0.0.1:{{add 8000 $index}}').set_protocol('trpc').set_serialization_type(
        SerializationType.PB).set_network('tcp').set_timeout(10000)
    {{- end }}
    ctx = TrpcContext()
    try:
            {{- if $rpc.ClientStreaming }}
        cstream: rpc_stream.{{$svrNameCamelCase}}ClientStream.{{$rpcNameCamelCase}} = await proxy.async{{$rpcName}}(ctx, options)
        assert cstream is not None
        for _ in range(1, 10):
            req = pb.{{$req}}()
            await cstream.send(req)

                {{- if $rpc.ServerStreaming }}

        await cstream.close_send()
                {{- else }}
        rsp = await cstream.close_and_recv()
        print(f"{{$rpcName}} close_and_recv rsp is : {rsp}")
                {{- end }}
            {{- else }}
        req = pb.{{$req}}()
        cstream: rpc_stream.{{$svrNameCamelCase}}ClientStream.{{$rpcNameCamelCase}} = await proxy.async{{$rpcName}}(ctx, req, options)
        assert cstream is not None
            {{- end }}
            {{- if $rpc.ServerStreaming }}

        while True:
            try:
                rsp = await cstream.recv()
                print(f"{{$rpcName}} recv rsp is : {rsp}")
            except EOFError as err:
                print(f"{{$rpcName}} recv done err: {err}")
                break
            {{- end }}
    except Exception as exc:  # pylint: disable=broad-exception-caught
        logger.error(exc)
        assert 0
async_test.append(test_{{$svrNameCamelCase}}_{{$rpcName}}())
        {{- end }}
    {{ end }}
{{- end}}

{{ if $legacy }}
if __name__ == "__main__":
    config_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "trpc_python.yaml"))
    trpc_config.load_global_config(config_path, 'utf-8')
    trpc_plugin.setup()

    for test_func in sync_test:
        test_func()
    loop = asyncio.get_event_loop()
    loop.run_until_complete(asyncio.gather(*async_test))
{{- else }}
async def run_all():
    """Run all async tests"""
    await asyncio.gather(*async_test)


if __name__ == "__main__":
    config_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "trpc_python.yaml"))
    TrpcConfig.load(config_path, 'utf-8')

    for test_func in sync_test:
        test_func()
    asyncio.run(run_all())
{{- end }}
