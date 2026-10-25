#!/bin/bash

python3 -m pip install --upgrade pip 

pip3 install yapf


# 查找并格式化 Python 文件
find . -name "*.py" -not -path "./venv/*" \
                    -not -path "./stub/*" \
                    -exec yapf -i {} \;
