#!/bin/sh
rm htmlcov/  -r  
rm *.xml
rm cov.tmp

unit_test_file_path=tests
report_file=test-unintest-report.xml
root_package=.

echo "generate coverage report"

python3 -m pytest  -rA -v --cov-report html --cov-report term --cov-report xml --cov=$root_package ${unit_test_file_path}/ --junitxml=$report_file > cov.tmp
