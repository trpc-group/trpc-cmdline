#!/bin/bash

PY3=python3


# 查找环境中最高版本的 python
function find_max_py(){
  # 构造python脚本来查找最高版本python比较方便
  py_data="
# coding=utf-8
import subprocess

def get_max_ver_python():
    \"\"\"\
    获取最大的python版本
    \"\"\"\

    pipe = subprocess.Popen('whereis python3', shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    res = pipe.stdout.read().decode()
    bins = res.split(\" \")[1:]
    mx_ver = \"python3\"
    max_ver = 6
    for bin in bins:
        py = bin.split(\"/\")[-1]
        if len(py.split(\".\")) < 2:
            continue
        minor = py.split(\".\")[1]
        if minor.isdigit() and int(minor) > max_ver:
            max_ver = int(minor)
            mx_ver = py
    return mx_ver

ver=get_max_ver_python()
print(ver)
"

    echo -e  "$py_data" > py_ver.py
    PY3=`python3 py_ver.py`
    rm -rf py_ver.py
}

find_max_py
echo "use pythons is ${PY3}"

#activate

function install_dependency_package() {
    echo "$1"
    if [ -f $1 ];then
         pip install -r $1
    fi
}

NEED_VENV=$1

if [ "$NEED_VENV" != "no_venv" ];then
    echo "need venv"
    if [ ! -d "venv" ];then
        $PY3 -m venv venv
    fi
    source venv/bin/activate
fi

$PY3 -m pip install --upgrade pip
pip install wheel
install_dependency_package "./requirements.txt"
install_dependency_package "./requirements/requirements-trpc.txt"
install_dependency_package "./requirements/requirements-test.txt"
