#!/bin/bash

sudo rm  -rf  coverage.xml
sudo rm  -rf  *.log
sudo rm  -rf  htmlcov
sudo rm  -rf  .coverage
sudo rm  -rf  .__*
sudo rm  -rf  *.egg-info
sudo rm  -rf dist/
sudo rm  -rf build/


sudo rm  -rf  test_tracemalloc*
sudo rm  -rf  test-ngtest-ut-*
sudo rm  -rf  cov.tmp
sudo rm  -rf  *.lock
sudo rm  -rf  agent/*.lock

sudo rm  -rf  .__py_trpc_frame.lock
sudo rm  -rf  .__trpc.lock

pip3 freeze > tmp_requirements.txt


find -type d | grep __pycache__ | xargs sudo rm  -r

find ./ -type f -name "*.log" -exec sudo rm {} \;


pip3 uninstall -r tmp_requirements.txt -y
sudo rm  -rf tmp_requirements.txt
