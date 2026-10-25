# Tencent is pleased to support the open source community by making tRPC-Python available.
#
# Copyright (C) 2026 Tencent. All rights reserved.
#
# tRPC-Python is licensed under Apache-2.0.
{{- $pkgName := .PackageName -}}
{{- $pbFile := trimright (last (splitList "." .Protofile)) .Protofile -}}
{{- $pbFile := trimright "." $pbFile -}}
{{- $pkgpath := gopkg $pkgName }}

import sys

from setuptools import find_packages
from setuptools import setup

if sys.version_info < (3, 6):
    raise RuntimeError("It requires Python 3.6+")

name = "{{$pkgpath}}"
version = "0.1.0"
packages = find_packages()

setup(
    name=name,
    version=version,
    packages=packages,
    author='',
    author_email='',
    maintainer='',
    maintainer_email='',
    url='',
    license='',
    description='stub package',
    package_data={'': ['*.proto']},

)
