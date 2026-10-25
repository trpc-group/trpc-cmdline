# Tencent is pleased to support the open source community by making tRPC-Python available.
#
# Copyright (C) 2026 Tencent. All rights reserved.
#
# tRPC-Python is licensed under Apache-2.0.

"""This is server module"""

{{- $legacy := eq .PythonSDK "trpc" -}}
{{- $pkgName := .PackageName -}}
{{- $pbFile := trimright (last (splitList "." .Protofile)) .Protofile -}}
{{- $pbFile := trimright "." $pbFile -}}
{{- $pkgpath := gopkg $pkgName }}

import os
import sys
import argparse
from typing import List
{{ if $legacy }}
import trpc
from trpc.log import logger
{{- else }}
from trpc_sdk.log import logger
from trpc_sdk.server import create_server
{{- end }}

# add the stub to the path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "stub")))  # noqa
if __name__ != '__main__':  # noqa
    sys.path.append(os.path.abspath(os.path.dirname(__file__)))  # noqa


# import sdk and plugins

# pylint: disable=unused-import

# 以下插件可以使用, 需要什么插件直接去除注释即可

# DO NOT remove the plugin below!
{{ if not $legacy }}# {{ end }}import trpc_metrics_runtime as _

from {{$pkgpath}} import rpc
{{range $service := .Services -}}
{{- $svrName := $service.Name -}}
{{- $svrNameCamelCase := $service.Name | camelcase }}
from {{snakecase $svrName}} import {{$svrNameCamelCase}}Servicer
{{- end}}

# argparse parser
parser = argparse.ArgumentParser(description='TRPC Python Framework Server')

DEFAULT_CONF_PATH = os.path.abspath(os.path.join(os.path.dirname(__file__), "trpc_python.yaml"))

# define config file argument
parser.add_argument(
    '--conf',
    '-c',
    default=DEFAULT_CONF_PATH,
    help='start server with conf, for example: `python3 trpc_main.py  --conf=trpc_python.yaml`'
)

def parse_args(argv: List[str]):
    """解析命令行参数
    :param argv: 命令行参数
    """
    try:
        args = parser.parse_args(argv[1:])  # Skip the script name
        if os.path.exists(args.conf):
            print(f"conf_path: {args.conf}")
            return args.conf
        print(f"conf_path: {args.conf} not exists")
        if os.path.exists(DEFAULT_CONF_PATH):
            print(f"use default conf path: {DEFAULT_CONF_PATH}")
            return DEFAULT_CONF_PATH
        print(f"default conf path: {DEFAULT_CONF_PATH} not exists, please check the path")
        sys.exit(1)
    except SystemExit:
        # argparse will call sys.exit() when --help is used or invalid args
        sys.exit(1)

def serve(conf_path: str):
    """启动服务
    :param conf_path: 配置文件路径
    """

    svr = {{if $legacy}}trpc.new{{else}}create_server{{end}}(conf_path)
{{range $service := .Services -}}
{{$svrName := $service.Name}}
{{- $svrNameCamelCase := $service.Name | camelcase }}
    rpc.register_{{$svrNameCamelCase}}Servicer_server(svr, {{$svrNameCamelCase}}Servicer())
{{- end}}

    svr.serve()


def main():
    """主函数"""

    # 解析命令行参数
    conf_path = parse_args(sys.argv)
    logger.info("load config %s", conf_path)

    # 启动服务
    serve(conf_path)

if __name__ == '__main__':
    sys.exit(main())
